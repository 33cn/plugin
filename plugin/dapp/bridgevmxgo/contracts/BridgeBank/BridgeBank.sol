pragma solidity ^0.5.0;

import "./GoAssetBank.sol";
import "./EvmAssetBank.sol";
import "../Oracle.sol";
import "../GoAssetBridge.sol";

/**
 * @title BridgeBank
 * @dev Bank contract which coordinates asset-related functionality.
 *      GoAssetBank manages the minting and burning of tokens which
 *      represent go contract issued assets, while EvmAssetBank manages
 *      the locking and unlocking of Chain33 and ERC20 token assets
 *      based on Chain33.
 **/

contract BridgeBank is GoAssetBank, EvmAssetBank {

    using SafeMath for uint256;

    address public operator;
    Oracle public oracle;
    GoAssetBridge public goAssetBridge;

    /*
    * @dev: Constructor, sets operator
    */
    constructor (
        address _operatorAddress,
        address _oracleAddress,
        address _goAssetBridgeAddress
    )
        public
    {
        operator = _operatorAddress;
        oracle = Oracle(_oracleAddress);
        goAssetBridge = GoAssetBridge(_goAssetBridgeAddress);
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
    * @dev: Modifier to restrict access to the goAsset bridge
    */
    modifier onlyGoAssetBridge()
    {
        require(
            msg.sender == address(goAssetBridge),
            "Access restricted to the goAsset bridge"
        );
        _;
    }
    /*
     * @dev: Modifier to make sure this symbol not created now
     */
    modifier onlyBridgeToken(address _token)
    {
        require(
            (address(0) != _token) && (msg.value == 0),
            "Only bridge token could be locked and tranfer to contract:evmxgo"
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
     * @dev: Mints new BankTokens
     *
     * @param _goAssetSender: The goAsset sender's address.
     * @param _chain33Recipient: The intended recipient's Chain33 address.
     * @param _bridgeTokenAddress: The bridge token address
     * @param _symbol: goAsset token symbol
     * @param _amount: number of goAsset tokens to be minted
     */
     function mintBridgeTokens(
        address  _goAssetSender,
        address payable _intendedRecipient,
        address _bridgeTokenAddress,
        string memory _symbol,
        uint256 _amount
    )
        public
        onlyGoAssetBridge
    {
        return mintNewBridgeTokens(
            _goAssetSender,
            _intendedRecipient,
            _bridgeTokenAddress,
            _symbol,
            _amount
        );
    }

    /*
     * @dev: Burns bank tokens
     *
     * @param _goAssetReceiver: The _goAsset receiver address in bytes.
     * @param _goAssetTokenAddress: The currency type
     * @param _amount: number of goAsset tokens to be burned
     */
    function burnBridgeTokens(address _goAssetReceiver, address _goAssetTokenAddress, uint256 _amount) public
    {
        return burnGoAssetTokens(
            msg.sender,
            _goAssetReceiver,
            _goAssetTokenAddress,
             _amount
        );
    }

    /*
    * @dev: addToken2LockList used to add token with the specified address to be
    *       allowed locked from GoAsset
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
        address _recipient,
        address _token,
        uint256 _amount
    )
        public
        availableNonce()
        onlyBridgeToken(_token)
        payable
    {
        string memory symbol;
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
        onlyGoAssetBridge
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
    function getGoAssetDepositStatus(
        bytes32 _id
    )
        public
        view
        returns(bool)
    {
        return isLockedGoAssetDeposit(_id);
    }

    /*
    * @dev: Allows access to a GoAsset deposit's information via its unique identifier.
    *
    * @param _id: The deposit to be viewed.
    * @return: Original sender's Chain33 address.
    * @return: Intended GoAsset recipient's address in bytes.
    * @return: The lock deposit's currency, denoted by a token address.
    * @return: The amount locked in the deposit.
    * @return: The deposit's unique nonce.
    */
    function viewGoAssetDeposit(
        bytes32 _id
    )
        public
        view
        returns(address, address payable, address, uint256)
    {
        return getGoAssetDeposit(_id);
    }

}