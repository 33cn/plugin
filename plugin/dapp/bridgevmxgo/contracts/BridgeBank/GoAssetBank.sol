pragma solidity ^0.5.0;

import "../../openzeppelin-solidity/contracts/math/SafeMath.sol";
import "./BridgeToken.sol";

/**
 * @title GoAssetBank
 * @dev Manages the deployment and minting of ERC20 compatible BridgeTokens
 *      which represent assets issued by Go contracts.
 * @dev 为chain33上的go合约发行的资产进行BridgeTokens合约部署，铸币和销毁
 **/

contract GoAssetBank {

    using SafeMath for uint256;

    uint256 public bridgeTokenCount;
    mapping(address => bool) public bridgeTokenWhitelist;
    mapping(bytes32 => bool) public bridgeTokenCreated;
    mapping(bytes32 => GoAssetDeposit) goAssetDeposits;
    mapping(bytes32 => GoAssetBurn) goAssetBurns;
    mapping(address => DepositBurnCount) depositBurnCounts;
    mapping(bytes32 => address) public token2address;

    struct GoAssetDeposit {
        address goAssetSender;
        address payable chain33Recipient;
        address bridgeTokenAddress;
        uint256 amount;
        bool exist;
        uint256 nonce;
    }

    struct DepositBurnCount {
        uint256 depositCount;
        uint256 burnCount;
    }

    struct GoAssetBurn {
        address goAssetSender;
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

    event LogGoAssetTokenBurn(
        address _token,
        string _symbol,
        uint256 _amount,
        address _ownerFrom,
        address _goAssetReceiver,
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
    * @dev: Creates a new GoAssetDeposit with a unique ID
    *
    * @param _goAssetSender: The _goAssetSender sender's address.
    * @param _chain33Recipient: The intended recipient's Chain33 address.
    * @param _token: The currency type
    * @param _amount: The amount in the deposit.
    * @return: The newly created GoAssetSenderDeposit's unique id.
    */
    function newGoAssetDeposit(
        address _goAssetSender,
        address payable _chain33Recipient,
        address _token,
        uint256 _amount
    )
        internal
        returns(bytes32)
    {
        DepositBurnCount memory depositBurnCount = depositBurnCounts[_token];
        depositBurnCount.depositCount = depositBurnCount.depositCount.add(1);
        depositBurnCounts[_token] = depositBurnCount;

        bytes32 depositID = keccak256(
            abi.encodePacked(
                _goAssetSender,
                _chain33Recipient,
                _token,
                _amount,
                depositBurnCount.depositCount
            )
        );

        goAssetDeposits[depositID] = GoAssetDeposit(
            _goAssetSender,
            _chain33Recipient,
            _token,
            _amount,
            true,
            depositBurnCount.depositCount
        );

        return depositID;
    }

    /*
    * @dev: Creates a new GoAssetBurn with a unique ID
        *
        * @param _goAssetSender: The go Asset Sender address
        * @param _chain33Owner: The owner's Chain33 address.
        * @param _token: The token Address
        * @param _amount: The amount to be burned.
        * @param _nonce: The nonce indicates the burn count for this token
        * @return: The newly created GoAssetBurn's unique id.
        */
        function newGoAssetBurn(
            address _goAssetSender,
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
                    _goAssetSender,
                    _chain33Owner,
                    _token,
                    _amount,
                    nonce
                )
            );

            goAssetBurns[burnID] = GoAssetBurn(
                _goAssetSender,
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
        depositBurnCounts[newBridgeTokenAddress] = DepositBurnCount(
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
     * @dev: Mints new goAsset tokens
     *
     * @param _goAssetSender:  The _goAssetSender sender's address.
     * @param _chain33Recipient: The intended recipient's Chain33 address.
     * @param _goAssetTokenAddress: The currency type
     * @param _symbol: goAsset token symbol
     * @param _amount: number of goAsset tokens to be minted
     */
     function mintNewBridgeTokens(
        address  _goAssetSender,
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

        newGoAssetDeposit(
            _goAssetSender,
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
     * @dev: Burn goAsset tokens
     *
     * @param _from: The address to be burned from
     * @param _goAssetReceiver: The receiver's GoAsset address in bytes.
     * @param _goAssetTokenAddress: The token address of goAsset asset issued on chain33
     * @param _amount: number of goAsset tokens to be minted
     */
    function burnGoAssetTokens(
        address payable _from,
        address _goAssetReceiver,
        address _goAssetTokenAddress,
        uint256 _amount
    )
        internal
    {
        // Must be whitelisted bridge token
        require(
            bridgeTokenWhitelist[_goAssetTokenAddress],
            "Token must be a whitelisted bridge token"
        );

        // burn bridge tokens
        BridgeToken bridgeTokenInstance = BridgeToken(_goAssetTokenAddress);
        bridgeTokenInstance.burnFrom(_from, _amount);

        DepositBurnCount memory depositBurnCount = depositBurnCounts[_goAssetTokenAddress];
        require(
            depositBurnCount.burnCount + 1 > depositBurnCount.burnCount,
            "burn nonce is not available"
        );
        depositBurnCount.burnCount = depositBurnCount.burnCount.add(1);
        depositBurnCounts[_goAssetTokenAddress] = depositBurnCount;

        newGoAssetBurn(
            _goAssetReceiver,
            _from,
            _goAssetTokenAddress,
            _amount,
            depositBurnCount.burnCount
        );

        emit LogGoAssetTokenBurn(
            _goAssetTokenAddress,
            bridgeTokenInstance.symbol(),
            _amount,
            _from,
            _goAssetReceiver,
            depositBurnCount.burnCount
        );
    }

    /*
    * @dev: Checks if an individual GoAssetDeposit exists.
    *
    * @param _id: The unique GoAssetDeposit's id.
    * @return: Boolean indicating if the GoAssetDeposit exists in memory.
    */
    function isLockedGoAssetDeposit(
        bytes32 _id
    )
        internal
        view
        returns(bool)
    {
        return(goAssetDeposits[_id].exist);
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
    function getGoAssetDeposit(
        bytes32 _id
    )
        internal
        view
        returns(address, address payable, address, uint256)
    {
        GoAssetDeposit memory deposit = goAssetDeposits[_id];

        return(
            deposit.goAssetSender,
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