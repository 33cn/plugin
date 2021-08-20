pragma solidity ^0.5.0;

import "../openzeppelin-solidity/contracts/math/SafeMath.sol";
import "./Valset.sol";
import "./BridgeBank/BridgeBank.sol";

contract EthereumBridge {

    using SafeMath for uint256;

    /*
    * @dev: Public variable declarations
    */
    address public operator;
    Valset public valset;
    address public oracle;
    bool public hasOracle;
    BridgeBank public bridgeBank;
    bool public hasBridgeBank;

    uint256 public prophecyClaimCount;
    mapping(bytes32 => ProphecyClaim) public prophecyClaims;

    enum Status {
        Null,
        Pending,
        Success,
        Failed
    }

    enum ClaimType {
        Unsupported,
        Burn,
        Lock
    }

    struct ProphecyClaim {
        ClaimType claimType;
        bytes ethereumSender;
        address payable chain33Receiver;
        address originalValidator;
        address tokenAddress;
        string symbol;
        uint256 amount;
        Status status;
    }

    /*
    * @dev: Event declarations
    */
    event LogOracleSet(
        address _oracle
    );

    event LogBridgeBankSet(
        address _bridgeBank
    );

    event LogNewProphecyClaim(
        uint256 _prophecyID,
        ClaimType _claimType,
        bytes _ethereumSender,
        address payable _chain33Receiver,
        address _validatorAddress,
        address _tokenAddress,
        string _symbol,
        uint256 _amount
    );

    event LogProphecyCompleted(
        bytes32 _claimID,
        ClaimType _claimType
    );

    /*
    * @dev: Modifier which only allows access to currently pending prophecies
    */
    modifier isPending(
        bytes32 _claimID
    )
    {
        require(
            isProphecyClaimActive(_claimID),
            "Prophecy claim is not active"
        );
        _;
    }

    /*
    * @dev: Modifier to restrict access to the operator.
    */
    modifier onlyOperator()
    {
        require(
            msg.sender == operator,
            'Must be the operator.'
        );
        _;
    }

    /*
    * @dev: Modifier to restrict access to the oracle.
    */
    modifier onlyOracle()
    {
        require(
            msg.sender == oracle,
            'Must be the oracle.'
        );
        _;
    }

      /*
    * @dev: The bridge is not active until oracle and bridge bank are set
    */
    modifier isActive()
    {
        require(
            hasOracle == true && hasBridgeBank == true,
            "The Operator must set the oracle and bridge bank for bridge activation"
        );
        _;
    }

    /*
    * @dev: Modifier to make sure the claim type is valid
    */
    modifier validClaimType(
        ClaimType _claimType
    )
    {
        require(
            _claimType == ClaimType.Burn || _claimType == ClaimType.Lock,
            "The claim type must be ClaimType.Burn or ClaimType.Lock"
        );
        _;
    }

    /*
    * @dev: Constructor
    */
    constructor(
        address _operator,
        address _valset
    )
        public
    {
        prophecyClaimCount = 0;
        operator = _operator;
        valset = Valset(_valset);
        hasOracle = false;
        hasBridgeBank = false;
    }

    /*
    * @dev: setOracle
    */
    function setOracle(
        address _oracle
    )
        public
        onlyOperator
    {
        require(
            !hasOracle,
            "The Oracle cannot be updated once it has been set"
        );

        hasOracle = true;
        oracle = _oracle;

        emit LogOracleSet(
            oracle
        );
    }

    /*
    * @dev: setBridgeBank
    */
    function setBridgeBank(
        address payable _bridgeBank
    )
        public
        onlyOperator
    {
        require(
            !hasBridgeBank,
            "The Bridge Bank cannot be updated once it has been set"
        );

        hasBridgeBank = true;
        bridgeBank = BridgeBank(_bridgeBank);

        emit LogBridgeBankSet(
            address(bridgeBank)
        );
    }

    /*
    * @dev: setNewProphecyClaim
    *       Sets a new burn or lock prophecy claim, adding it to the prophecyClaims mapping.
    *       Lock claims can only be created for BridgeTokens on BridgeBank's whitelist. The operator
    *       is responsible for adding them, and lock claims will fail until the operator has done so.
    */
    function setNewProphecyClaim(
        bytes32 _claimID,
        uint8 _claimType,
        bytes memory _ethereumSender,
        address payable _chain33Receiver,
        address _originalValidator,
        address _tokenAddress,
        string memory _symbol,
        uint256 _amount
    )
        public
        isActive
        onlyOracle
    {
        // Increment the prophecy claim count
        prophecyClaimCount = prophecyClaimCount.add(1);
        ClaimType claimType = ClaimType(_claimType);

        //overwrite the token address in case of lock
        if (claimType == ClaimType.Lock) {
            _tokenAddress = bridgeBank.getToken2address(_symbol);
        }

        // Create the new ProphecyClaim
        ProphecyClaim memory prophecyClaim = ProphecyClaim(
            claimType,
            _ethereumSender,
            _chain33Receiver,
            _originalValidator,
            _tokenAddress,
            _symbol,
            _amount,
            Status.Pending
        );

        // Add the new ProphecyClaim to the mapping
        prophecyClaims[_claimID] = prophecyClaim;

        emit LogNewProphecyClaim(
            prophecyClaimCount,
            claimType,
            _ethereumSender,
            _chain33Receiver,
            _originalValidator,
            _tokenAddress,
            _symbol,
            _amount
        );
    }

    /*
    * @dev: completeClaim
    *       Allows for the completion of ProphecyClaims once processed by the Oracle.
    *       Burn claims unlock tokens stored by BridgeBank.
    *       Lock claims mint BridgeTokens on BridgeBank's token whitelist.
    */
    function completeClaim(
        bytes32 _claimID
    )
        public
        isPending(_claimID)
    {
        require(
            msg.sender == oracle,
            "Only the Oracle may complete prophecies"
        );

        prophecyClaims[_claimID].status = Status.Success;

        ClaimType claimType = prophecyClaims[_claimID].claimType;
        if(claimType == ClaimType.Burn) {
            unlockTokens(_claimID);
        } else {
            issueBridgeTokens(_claimID);
        }

        emit LogProphecyCompleted(
            _claimID,
            claimType
        );
    }

    /*
    * @dev: issueBridgeTokens
    *       Issues a request for the BridgeBank to mint new BridgeTokens
    */
    function issueBridgeTokens(
        bytes32 _claimID
    )
        internal
    {
        ProphecyClaim memory prophecyClaim = prophecyClaims[_claimID];

        bridgeBank.mintBridgeTokens(
            prophecyClaim.ethereumSender,
            prophecyClaim.chain33Receiver,
            prophecyClaim.tokenAddress,
            prophecyClaim.symbol,
            prophecyClaim.amount
        );
    }

    /*
    * @dev: unlockTokens
    *       Issues a request for the BridgeBank to unlock funds held on contract
    */
    function unlockTokens(
        bytes32 _claimID
    )
        internal
    {
        ProphecyClaim memory prophecyClaim = prophecyClaims[_claimID];

        bridgeBank.unlock(
            prophecyClaim.chain33Receiver,
            prophecyClaim.tokenAddress,
            prophecyClaim.symbol,
            prophecyClaim.amount
        );
    }

    /*
    * @dev: isProphecyClaimActive
    *       Returns boolean indicating if the ProphecyClaim is active
    */
    function isProphecyClaimActive(
        bytes32 _claimID
    )
        public
        view
        returns(bool)
    {
        return prophecyClaims[_claimID].status == Status.Pending;
    }

    /*
    * @dev: isProphecyValidatorActive
    *       Returns boolean indicating if the validator that originally
    *       submitted the ProphecyClaim is still an active validator
    */
    function isProphecyClaimValidatorActive(
        bytes32 _claimID
    )
        public
        view
        returns(bool)
    {
        return valset.isActiveValidator(
            prophecyClaims[_claimID].originalValidator
        );
    }

    /*
    * @dev: Modifier to make sure the claim type is valid
    */
    function isValidClaimType(uint8 _claimType) public pure returns(bool)
    {
        ClaimType claimType = ClaimType(_claimType);
        if (claimType == ClaimType.Burn || claimType == ClaimType.Lock) {
            return true;
        }
        return false;
    }
}
