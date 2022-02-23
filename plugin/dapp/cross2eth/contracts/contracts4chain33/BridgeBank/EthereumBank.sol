pragma solidity ^0.5.0;

import "../../openzeppelin-solidity/contracts/math/SafeMath.sol";
import "./BridgeToken.sol";

/**
 * @title EthereumBank
 * @dev Manages the deployment and minting of ERC20 compatible BridgeTokens
 *      which represent assets based on the Ethereum blockchain.
 **/

contract EthereumBank {

    using SafeMath for uint256;

    uint256 public bridgeTokenCount;
    address payable public proxyReceiver;
    mapping(address => bool) public bridgeTokenWhitelist;
    mapping(bytes32 => bool) public bridgeTokenCreated;
    mapping(bytes32 => EthereumDeposit) ethereumDeposits;
    mapping(bytes32 => EthereumBurn) ethereumBurns;
    mapping(address => DepositBurnWithdrawCount) depositBurnWithdrawCounts;
    mapping(bytes32 => address) public token2address;

    struct EthereumDeposit {
        bytes ethereumSender;
        address payable chain33Recipient;
        address bridgeTokenAddress;
        uint256 amount;
        bool exist;
        uint256 nonce;
    }

    struct DepositBurnWithdrawCount {
        uint256 depositCount;
        uint256 burnCount;
        uint256 withdrawCount;

    }

    struct EthereumBurn {
        bytes ethereumSender;
        address payable chain33Owner;
        address bridgeTokenAddress;
        uint256 amount;
        uint256 nonce;
    }

    /*
    * @dev: Event declarations
    */
    event LogNewBridgeToken(
        address _token,
        string _symbol
    );

    event LogBridgeTokenMint(
        address _token,
        string _symbol,
        uint256 _amount,
        address _beneficiary
    );

    event LogEthereumTokenBurn(
        address _token,
        string _symbol,
        uint256 _amount,
        address _ownerFrom,
        bytes _ethereumReceiver,
        uint256 _nonce
    );

    event LogEthereumTokenWithdraw(
        address _bridgeToken,
        string _symbol,
        uint256 _amount,
        address _ownerFrom,
        bytes _ethereumReceiver,
        address _proxyReceiver,
        uint256 _nonce
    );

    /*
     * @dev: Modifier to make sure this symbol not created now
     */
     modifier notCreated(string memory _symbol)
     {
         require(
             !hasBridgeTokenCreated(_symbol),
             "The symbol has been created already"
         );
         _;
     }

     /*
     * @dev: Modifier to make sure this symbol not created now
     */
     modifier created(string memory _symbol)
     {
         require(
             hasBridgeTokenCreated(_symbol),
             "The symbol has not been created yet"
         );
         _;
     }

    /*
    * @dev: Constructor, sets bridgeTokenCount
    */
    constructor () public {
        bridgeTokenCount = 0;
    }

    /*
    * @dev: check whether this symbol has been created yet or not
    *
    * @param _symbol: token symbol
    * @return: true or false
    */
    function hasBridgeTokenCreated(string memory _symbol) public view returns(bool) {
        bytes32 symHash = keccak256(abi.encodePacked(_symbol));
        return bridgeTokenCreated[symHash];
    }

    /*
    * @dev: Creates a new EthereumDeposit with a unique ID
    *
    * @param _ethereumSender: The sender's Ethereum address in bytes.
    * @param _chain33Recipient: The intended recipient's Chain33 address.
    * @param _token: The currency type
    * @param _amount: The amount in the deposit.
    * @return: The newly created EthereumDeposit's unique id.
    */
    function newEthereumDeposit(
        bytes memory _ethereumSender,
        address payable _chain33Recipient,
        address _token,
        uint256 _amount
    )
        internal
        returns(bytes32)
    {
        DepositBurnWithdrawCount memory depositBurnCount = depositBurnWithdrawCounts[_token];
        depositBurnCount.depositCount = depositBurnCount.depositCount.add(1);
        depositBurnWithdrawCounts[_token] = depositBurnCount;

        bytes32 depositID = keccak256(
            abi.encodePacked(
                _ethereumSender,
                _chain33Recipient,
                _token,
                _amount,
                depositBurnCount.depositCount
            )
        );

        ethereumDeposits[depositID] = EthereumDeposit(
            _ethereumSender,
            _chain33Recipient,
            _token,
            _amount,
            true,
            depositBurnCount.depositCount
        );

        return depositID;
    }

    /*
    * @dev: Creates a new EthereumBurn with a unique ID
        *
        * @param _ethereumSender: The sender's Ethereum address in bytes.
        * @param _chain33Owner: The owner's Chain33 address.
        * @param _token: The token Address
        * @param _amount: The amount to be burned.
        * @param _nonce: The nonce indicates the burn count for this token
        * @return: The newly created EthereumBurn's unique id.
        */
        function newEthereumBurn(
            bytes memory _ethereumSender,
            address payable _chain33Owner,
            address _token,
            uint256 _amount,
            uint256 nonce
        )
            internal
            returns(bytes32)
        {
            bytes32 burnID = keccak256(
                abi.encodePacked(
                    _ethereumSender,
                    _chain33Owner,
                    _token,
                    _amount,
                    nonce
                )
            );

            ethereumBurns[burnID] = EthereumBurn(
                _ethereumSender,
                _chain33Owner,
                _token,
                _amount,
                nonce
            );

            return burnID;
        }


    /*
     * @dev: Deploys a new BridgeToken contract
     *
     * @param _symbol: The BridgeToken's symbol
     */
    function deployNewBridgeToken(
        string memory _symbol
    )
        internal
        notCreated(_symbol)
        returns(address)
    {
        bridgeTokenCount = bridgeTokenCount.add(1);

        // Deploy new bridge token contract
        BridgeToken newBridgeToken = (new BridgeToken)(_symbol);

        // Set address in tokens mapping
        address newBridgeTokenAddress = address(newBridgeToken);
        bridgeTokenWhitelist[newBridgeTokenAddress] = true;
        bytes32 symHash = keccak256(abi.encodePacked(_symbol));
        bridgeTokenCreated[symHash] = true;
        depositBurnWithdrawCounts[newBridgeTokenAddress] = DepositBurnWithdrawCount(
            uint256(0),
            uint256(0),
            uint256(0));
        token2address[symHash] = newBridgeTokenAddress;

        emit LogNewBridgeToken(
            newBridgeTokenAddress,
            _symbol
        );

        return newBridgeTokenAddress;
    }

    /*
     * @dev: Mints new ethereum tokens
     *
     * @param _ethereumSender: The sender's Ethereum address in bytes.
     * @param _chain33Recipient: The intended recipient's Chain33 address.
     * @param _ethereumTokenAddress: The currency type
     * @param _symbol: ethereum token symbol
     * @param _amount: number of ethereum tokens to be minted
     */
     function mintNewBridgeTokens(
        bytes memory _ethereumSender,
        address payable _intendedRecipient,
        address _bridgeTokenAddress,
        string memory _symbol,
        uint256 _amount
    )
        internal
    {
        // Must be whitelisted bridge token
        require(
            bridgeTokenWhitelist[_bridgeTokenAddress],
            "Token must be a whitelisted bridge token"
        );

        // Mint bridge tokens
        require(
            BridgeToken(_bridgeTokenAddress).mint(
                _intendedRecipient,
                _amount
            ),
            "Attempted mint of bridge tokens failed"
        );

        newEthereumDeposit(
            _ethereumSender,
            _intendedRecipient,
            _bridgeTokenAddress,
            _amount
        );

        emit LogBridgeTokenMint(
            _bridgeTokenAddress,
            _symbol,
            _amount,
            _intendedRecipient
        );
    }

    /*
     * @dev: Burn ethereum tokens
     *
     * @param _from: The address to be burned from
     * @param _ethereumReceiver: The receiver's Ethereum address in bytes.
     * @param _ethereumTokenAddress: The token address of ethereum asset issued on chain33
     * @param _amount: number of ethereum tokens to be burned
     */
    function burnEthereumTokens(
        address payable _from,
        bytes memory _ethereumReceiver,
        address _ethereumTokenAddress,
        uint256 _amount
    )
        internal
    {
        // Must be whitelisted bridge token
        require(
            bridgeTokenWhitelist[_ethereumTokenAddress],
            "Token must be a whitelisted bridge token"
        );

        // burn bridge tokens
        BridgeToken bridgeTokenInstance = BridgeToken(_ethereumTokenAddress);
        bridgeTokenInstance.burnFrom(_from, _amount);

        DepositBurnWithdrawCount memory depositBurnCount = depositBurnWithdrawCounts[_ethereumTokenAddress];
        require(
            depositBurnCount.burnCount + 1 > depositBurnCount.burnCount,
            "burn nonce is not available"
        );
        depositBurnCount.burnCount = depositBurnCount.burnCount.add(1);
        depositBurnWithdrawCounts[_ethereumTokenAddress] = depositBurnCount;

        newEthereumBurn(
            _ethereumReceiver,
            _from,
            _ethereumTokenAddress,
            _amount,
            depositBurnCount.burnCount
        );

        emit LogEthereumTokenBurn(
            _ethereumTokenAddress,
            bridgeTokenInstance.symbol(),
            _amount,
            _from,
            _ethereumReceiver,
            depositBurnCount.burnCount
        );
    }

    /*
     * @dev:  withdraw ethereum tokens
     *
     * @param _from: The address to be withdrew from
     * @param _ethereumReceiver: The receiver's Ethereum address in bytes.
     * @param _bridgeTokenAddress: The token address of ethereum asset issued on chain33
     * @param _amount: number of ethereum tokens to be withdrew
     */
    function withdrawEthereumTokens(
        address payable _from,
        bytes memory _ethereumReceiver,
        address _bridgeTokenAddress,
        uint256 _amount
    )
    internal
    {
        require(proxyReceiver != address(0), "proxy receiver hasn't been set");
        // Must be whitelisted bridge token
        require(bridgeTokenWhitelist[_bridgeTokenAddress], "Token must be a whitelisted bridge token");
        // burn bridge tokens
        BridgeToken bridgeTokenInstance = BridgeToken(_bridgeTokenAddress);
        bridgeTokenInstance.transferFrom(_from, proxyReceiver, _amount);

        DepositBurnWithdrawCount memory wdCount = depositBurnWithdrawCounts[_bridgeTokenAddress];
        require(
            wdCount.withdrawCount + 1 > wdCount.withdrawCount,
            "withdraw nonce is not available"
        );
        wdCount.withdrawCount = wdCount.withdrawCount.add(1);
        depositBurnWithdrawCounts[_bridgeTokenAddress] = wdCount;

        emit LogEthereumTokenWithdraw(
            _bridgeTokenAddress,
            bridgeTokenInstance.symbol(),
            _amount,
            _from,
            _ethereumReceiver,
            proxyReceiver,
            wdCount.withdrawCount
        );

    }

    /*
    * @dev: Checks if an individual EthereumDeposit exists.
    *
    * @param _id: The unique EthereumDeposit's id.
    * @return: Boolean indicating if the EthereumDeposit exists in memory.
    */
    function isLockedEthereumDeposit(
        bytes32 _id
    )
        internal
        view
        returns(bool)
    {
        return(ethereumDeposits[_id].exist);
    }

  /*
    * @dev: Gets an item's information
    *
    * @param _Id: The item containing the desired information.
    * @return: Sender's address.
    * @return: Recipient's address in bytes.
    * @return: Token address.
    * @return: Amount of chain33/erc20 in the item.
    * @return: Unique nonce of the item.
    */
    function getEthereumDeposit(
        bytes32 _id
    )
        internal
        view
        returns(bytes memory, address payable, address, uint256)
    {
        EthereumDeposit memory deposit = ethereumDeposits[_id];

        return(
            deposit.ethereumSender,
            deposit.chain33Recipient,
            deposit.bridgeTokenAddress,
            deposit.amount
        );
    }

    function getToken2address(string memory _symbol)
        created(_symbol)
        public view returns(address)
    {
        bytes32 symHash = keccak256(abi.encodePacked(_symbol));
        return token2address[symHash];
    }
}