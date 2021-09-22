pragma solidity ^0.5.0;

import "../../openzeppelin-solidity/contracts/math/SafeMath.sol";
import "./BridgeToken.sol";

/**
 * @title Chain33Bank
 * @dev Manages the deployment and minting of ERC20 compatible BridgeTokens
 *      which represent assets based on the Chain33 blockchain.
 **/

contract Chain33Bank {

    using SafeMath for uint256;

    uint256 public bridgeTokenCount;
    mapping(address => bool) public bridgeTokenWhitelist;
    mapping(bytes32 => bool) public bridgeTokenCreated;
    mapping(bytes32 => Chain33Deposit) chain33Deposits;
    mapping(bytes32 => Chain33Burn) chain33Burns;
    mapping(address => DepositBurnCount) depositBurnCounts;
    mapping(bytes32 => address) public token2address;

    struct Chain33Deposit {
        bytes chain33Sender;
        address payable ethereumRecipient;
        address bridgeTokenAddress;
        uint256 amount;
        bool exist;
        uint256 nonce;
    }

    struct DepositBurnCount {
        uint256 depositCount;
        uint256 burnCount;
    }

    struct Chain33Burn {
        bytes chain33Sender;
        address payable ethereumOwner;
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

    event LogChain33TokenBurn(
        address _token,
        string _symbol,
        uint256 _amount,
        address _ownerFrom,
        bytes _chain33Receiver,
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
    * @dev: Creates a new Chain33Deposit with a unique ID
    *
    * @param _chain33Sender: The sender's Chain33 address in bytes.
    * @param _ethereumRecipient: The intended recipient's Ethereum address.
    * @param _token: The currency type
    * @param _amount: The amount in the deposit.
    * @return: The newly created Chain33Deposit's unique id.
    */
    function newChain33Deposit(
        bytes memory _chain33Sender,
        address payable _ethereumRecipient,
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
                _chain33Sender,
                _ethereumRecipient,
                _token,
                _amount,
                depositBurnCount.depositCount
            )
        );

        chain33Deposits[depositID] = Chain33Deposit(
            _chain33Sender,
            _ethereumRecipient,
            _token,
            _amount,
            true,
            depositBurnCount.depositCount
        );

        return depositID;
    }

    /*
    * @dev: Creates a new Chain33Burn with a unique ID
        *
        * @param _chain33Sender: The sender's Chain33 address in bytes.
        * @param _ethereumOwner: The owner's Ethereum address.
        * @param _token: The token Address
        * @param _amount: The amount to be burned.
        * @param _nonce: The nonce indicates the burn count for this token
        * @return: The newly created Chain33Burn's unique id.
        */
        function newChain33Burn(
            bytes memory _chain33Sender,
            address payable _ethereumOwner,
            address _token,
            uint256 _amount,
            uint256 nonce
        )
            internal
            returns(bytes32)
        {
            bytes32 burnID = keccak256(
                abi.encodePacked(
                    _chain33Sender,
                    _ethereumOwner,
                    _token,
                    _amount,
                    nonce
                )
            );

            chain33Burns[burnID] = Chain33Burn(
                _chain33Sender,
                _ethereumOwner,
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
     * @dev: Mints new chain33 tokens
     *
     * @param _chain33Sender: The sender's Chain33 address in bytes.
     * @param _ethereumRecipient: The intended recipient's Ethereum address.
     * @param _chain33TokenAddress: The currency type
     * @param _symbol: chain33 token symbol
     * @param _amount: number of chain33 tokens to be minted
     */
     function mintNewBridgeTokens(
        bytes memory _chain33Sender,
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

        newChain33Deposit(
            _chain33Sender,
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
     * @dev: Burn chain33 tokens
     *
     * @param _from: The address to be burned from
     * @param _chain33Receiver: The receiver's Chain33 address in bytes.
     * @param _chain33TokenAddress: The token address of chain33 asset issued on ethereum
     * @param _amount: number of chain33 tokens to be minted
     */
    function burnChain33Tokens(
        address payable _from,
        bytes memory _chain33Receiver,
        address _chain33TokenAddress,
        uint256 _amount
    )
        internal
    {
        // Must be whitelisted bridge token
        require(
            bridgeTokenWhitelist[_chain33TokenAddress],
            "Token must be a whitelisted bridge token"
        );

        // burn bridge tokens
        BridgeToken bridgeTokenInstance = BridgeToken(_chain33TokenAddress);
        bridgeTokenInstance.burnFrom(_from, _amount);

        DepositBurnCount memory depositBurnCount = depositBurnCounts[_chain33TokenAddress];
        require(
            depositBurnCount.burnCount + 1 > depositBurnCount.burnCount,
            "burn nonce is not available"
        );
        depositBurnCount.burnCount = depositBurnCount.burnCount.add(1);
        depositBurnCounts[_chain33TokenAddress] = depositBurnCount;

        newChain33Burn(
            _chain33Receiver,
            _from,
            _chain33TokenAddress,
            _amount,
            depositBurnCount.burnCount
        );

        emit LogChain33TokenBurn(
            _chain33TokenAddress,
            bridgeTokenInstance.symbol(),
            _amount,
            _from,
            _chain33Receiver,
            depositBurnCount.burnCount
        );
    }

    /*
    * @dev: Checks if an individual Chain33Deposit exists.
    *
    * @param _id: The unique Chain33Deposit's id.
    * @return: Boolean indicating if the Chain33Deposit exists in memory.
    */
    function isLockedChain33Deposit(
        bytes32 _id
    )
        internal
        view
        returns(bool)
    {
        return(chain33Deposits[_id].exist);
    }

  /*
    * @dev: Gets an item's information
    *
    * @param _Id: The item containing the desired information.
    * @return: Sender's address.
    * @return: Recipient's address in bytes.
    * @return: Token address.
    * @return: Amount of ethereum/erc20 in the item.
    * @return: Unique nonce of the item.
    */
    function getChain33Deposit(
        bytes32 _id
    )
        internal
        view
        returns(bytes memory, address payable, address, uint256)
    {
        Chain33Deposit memory deposit = chain33Deposits[_id];

        return(
            deposit.chain33Sender,
            deposit.ethereumRecipient,
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

    function getToken2addressV2(string memory _symbol)
        created(_symbol)
        public view returns(address, bool)
    {
        bytes32 symHash = keccak256(abi.encodePacked(_symbol));
        if (true != bridgeTokenCreated[symHash]) {
            return (address(0), false);
        }
        return (token2address[symHash], true);
    }
}