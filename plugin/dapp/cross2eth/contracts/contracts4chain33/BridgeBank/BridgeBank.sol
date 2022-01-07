pragma solidity ^0.5.0;

import "./EthereumBank.sol";
import "./Chain33Bank.sol";
import "../Oracle.sol";
import "../EthereumBridge.sol";

/**
 * @title BridgeBank
 * @dev Bank contract which coordinates asset-related functionality.
 *      EthereumBank manages the minting and burning of tokens which
 *      represent Ethereum based assets, while Chain33Bank manages
 *      the locking and unlocking of Chain33 and ERC20 token assets
 *      based on Chain33.
 **/

contract BridgeBank is EthereumBank, Chain33Bank {

    using SafeMath for uint256;
    
    address public operator;
    Oracle public oracle;
    EthereumBridge public ethereumBridge;

    /*
    * @dev: Constructor, sets operator
    */
    constructor (
        address _operatorAddress,
        address _oracleAddress,
        address _ethereumBridgeAddress
    )
        public
    {
        operator = _operatorAddress;
        oracle = Oracle(_oracleAddress);
        ethereumBridge = EthereumBridge(_ethereumBridgeAddress);
    }

    /*
    * @dev: Modifier to restrict access to operator
    */
    modifier onlyOperator() {
        require(
            msg.sender == operator,
            'Must be BridgeBank operator.'
        );
        _;
    }

    /*
     * @dev: Modifier to restrict access to Offline
     */
    modifier onlyOffline() {
        require(
            msg.sender == offlineSave,
            'Must be onlyOffline.'
        );
        _;
    }

    /*
    * @dev: Modifier to restrict access to the oracle
    */
    modifier onlyOracle()
    {
        require(
            msg.sender == address(oracle),
            "Access restricted to the oracle"
        );
        _;
    }

    /*
    * @dev: Modifier to restrict access to the ethereum bridge
    */
    modifier onlyEthereumBridge()
    {
        require(
            msg.sender == address(ethereumBridge),
            "Access restricted to the ethereum bridge"
        );
        _;
    }

   /*
    * @dev: Fallback function allows operator to send funds to the bank directly
    *       This feature is used for testing and is available at the operator's own risk.
    */
    function() external payable onlyOffline {}

    /*
    * @dev: Creates a new BridgeToken
    *
    * @param _symbol: The new BridgeToken's symbol
    * @return: The new BridgeToken contract's address
    */
    function createNewBridgeToken(
        string memory _symbol
    )
        public
        onlyOperator
        returns(address)
    {
        return deployNewBridgeToken(_symbol);
    }

    /*
    * @dev: set a proxy address to receive and it's transfer asset on Ethereum
    *
    * @param _proxyReceiver: The address to receive asset
    * @return: indicate whether set successfully or not
    */
    function setWithdrawProxy(address payable _proxyReceiver) public onlyOperator
    {
        proxyReceiver = _proxyReceiver;
    }

    /*
     * @dev: Mints new BankTokens
     *
     * @param _ethereumSender: The sender's Ethereum address in bytes.
     * @param _chain33Recipient: The intended recipient's Chain33 address.
     * @param _ethereumTokenAddress: The currency type
     * @param _symbol: ethereum token symbol
     * @param _amount: number of ethereum tokens to be minted
     */
     function mintBridgeTokens(
        bytes memory _ethereumSender,
        address payable _intendedRecipient,
        address _bridgeTokenAddress,
        string memory _symbol,
        uint256 _amount
    )
        public
        onlyEthereumBridge
    {
        return mintNewBridgeTokens(
            _ethereumSender,
            _intendedRecipient,
            _bridgeTokenAddress,
            _symbol,
            _amount
        );
    }

    /*
     * @dev: Burns bank tokens
     *
     * @param _ethereumReceiver: The _ethereum receiver address in bytes.
     * @param _ethereumTokenAddress: The token address mint on chain33 and it's origin from Ethereum
     * @param _amount: number of ethereum tokens to be burned
     */
    function burnBridgeTokens(
        bytes memory _ethereumReceiver,
        address _ethereumTokenAddress,
        uint256 _amount
    )
        public
    {
        return burnEthereumTokens(
            msg.sender,
            _ethereumReceiver,
            _ethereumTokenAddress,
             _amount
        );
    }

    /*
     * @dev: withdraw asset via Proxy
     *
     * @param _ethereumReceiver: The _ethereum receiver address in bytes.
     * @param _bridgeTokenAddress: The bridge Token Address issued in chain33 and it's origin from Ethereum/BSC
     * @param _amount: number of bridge tokens to be transferred to proxy address
     */
    function withdrawViaProxy(
        bytes memory _ethereumReceiver,
        address _bridgeTokenAddress,
        uint256 _amount
    )
    public
    {
        return withdrawEthereumTokens(
            msg.sender,
            _ethereumReceiver,
            _bridgeTokenAddress,
            _amount
        );
    }

    /*
    * @dev: addToken2LockList used to add token with the specified address to be
    *       allowed locked from Ethereum
    *
    * @param _token: token contract address
    * @param _symbol: token symbol
    */
    function addToken2LockList(
        address _token,
        string memory _symbol
    )
        public
        onlyOperator
    {
        addToken2AllowLock(_token, _symbol);
    }

   /*
    * @dev: configTokenOfflineSave used to config threshold to trigger tranfer token to offline account
    *       when the balance of locked token reaches
    *
    * @param _token: token contract address
    * @param _symbol:token symbol,just used for double check that token address and symbol is consistent
    * @param _threshold: _threshold to trigger transfer
    * @param _percents: amount to transfer per percents of threshold
    */
    function configLockedTokenOfflineSave(
        address _token,
        string memory _symbol,
        uint256 _threshold,
        uint8 _percents
    )
    public
    onlyOperator
    {
        if (address(0) != _token) {
            require(keccak256(bytes(BridgeToken(_token).symbol())) == keccak256(bytes(_symbol)), "token address and symbol is not consistent");
        } else {
            require(keccak256(bytes("BTY")) == keccak256(bytes(_symbol)), "token address and symbol is not consistent");
        }

        configOfflineSave4Lock(_token, _symbol, _threshold, _percents);
    }

   /*
   * @dev: configOfflineSaveAccount used to config offline account to receive token
   *       when the balance of locked token reaches threshold
   *
   * @param _offlineSave: receiver address
   */
    function configOfflineSaveAccount(address payable _offlineSave) public onlyOperator
    {
        offlineSave = _offlineSave;
    }

    /*
    * @dev: Locks received Chain33 funds.
    *
    * @param _recipient: bytes representation of destination address.
    * @param _token: token address in origin chain (0x0 if chain33)
    * @param _amount: value of deposit
    */
    function lock(
        bytes memory _recipient,
        address _token,
        uint256 _amount
    )
        public
        availableNonce()
        payable
    {
        string memory symbol;

        // Chain33 deposit
        if (msg.value > 0) {
          require(
              _token == address(0),
              "BTY deposits require the 'token' address to be the null address"
            );
          require(
              msg.value == _amount,
              "The transactions value must be equal the specified amount(BTY decimal is 8)"
            );

          // Set the the symbol to BTY
          symbol = "BTY";
          // ERC20 deposit
        } else {
          require(
              BridgeToken(_token).transferFrom(msg.sender, address(this), _amount),
              "Contract token allowances insufficient to complete this lock request"
          );
          // Set symbol to the ERC20 token's symbol
          symbol = BridgeToken(_token).symbol();
          require(
              tokenAllow2Lock[keccak256(abi.encodePacked(symbol))] == _token,
              'The token is not allowed to be locked from Chain33.'
          );
        }

        lockFunds(
            msg.sender,
            _recipient,
            _token,
            symbol,
            _amount
        );
    }

   /*
    * @dev: Unlocks Chain33 and ERC20 tokens held on the contract.
    *
    * @param _recipient: recipient's Chain33 address
    * @param _token: token contract address
    * @param _symbol: token symbol
    * @param _amount: wei amount or ERC20 token count
\   */
     function unlock(
        address payable _recipient,
        address _token,
        string memory _symbol,
        uint256 _amount
    )
        public
        onlyEthereumBridge
        hasLockedFunds(
            _token,
            _amount
        )
        canDeliver(
            _token,
            _amount
        )
    {
        unlockFunds(
            _recipient,
            _token,
            _symbol,
            _amount
        );
    }

    /*
    * @dev: Exposes an item's current status.
    *
    * @param _id: The item in question.
    * @return: Boolean indicating the lock status.
    */
    function getEthereumDepositStatus(
        bytes32 _id
    )
        public
        view
        returns(bool)
    {
        return isLockedEthereumDeposit(_id);
    }

    /*
    * @dev: Allows access to a Ethereum deposit's information via its unique identifier.
    *
    * @param _id: The deposit to be viewed.
    * @return: Original sender's Chain33 address.
    * @return: Intended Ethereum recipient's address in bytes.
    * @return: The lock deposit's currency, denoted by a token address.
    * @return: The amount locked in the deposit.
    * @return: The deposit's unique nonce.
    */
    function viewEthereumDeposit(
        bytes32 _id
    )
        public
        view
        returns(bytes memory, address payable, address, uint256)
    {
        return getEthereumDeposit(_id);
    }

}