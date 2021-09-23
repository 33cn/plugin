pragma solidity ^0.5.0;

import "../openzeppelin-solidity/contracts/math/SafeMath.sol";
import "./Valset.sol";
import "./EthereumBridge.sol";

contract Oracle {

    using SafeMath for uint256;

    /*
    * @dev: Public variable declarations
    */
    EthereumBridge public ethereumBridge;
    Valset public valset;
    address public operator;

    // Tracks the number of OracleClaims made on an individual BridgeClaim
    mapping(bytes32 => address[]) public oracleClaimValidators;
    mapping(bytes32 => mapping(address => bool)) public hasMadeClaim;

    enum ClaimType {
        Unsupported,
        Burn,
        Lock
    }

    /*
    * @dev: Event declarations
    */
    event LogNewOracleClaim(
        bytes32 _claimID,
        address _validatorAddress,
        bytes _signature
    );

    event LogProphecyProcessed(
        bytes32 _claimID,
        uint256 _weightedSignedPower,
        uint256 _weightedTotalPower,
        address _submitter
    );

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
    * @dev: Modifier to restrict access to current ValSet validators
    */
    modifier onlyValidator()
    {
        require(
            valset.isActiveValidator(msg.sender),
            "Must be an active validator"
        );
        _;
    }

    /*
    * @dev: Modifier to restrict access to current ValSet validators
    */
    modifier isPending(
        bytes32 _claimID
    )
    {
        require(
            ethereumBridge.isProphecyClaimActive(
                _claimID
            ) == true,
            "The prophecy must be pending for this operation"
        );
        _;
    }

    /*
    * @dev: Modifier to restrict the claim type must be burn or lock
    */
    modifier isValidClaimType(
        ClaimType _claimType
    )
    {
        require(
           ethereumBridge.isValidClaimType(
               uint8(_claimType)
           ) == true,
           "The claim type must be burn or lock"
        );
            _;
        }

    /*
    * @dev: Constructor
    */
    constructor(
        address _operator,
        address _valset,
        address _ethereumBridge
    )
        public
    {
        operator = _operator;
        ethereumBridge = EthereumBridge(_ethereumBridge);
        valset = Valset(_valset);
    }

    /*
    * @dev: newOracleClaim
    *       Allows validators to make new OracleClaims on ethereum lock/burn prophecy,
    *       if the required vote power reached,just make it processed
    * @param _claimType: burn or lock,
    * @param _ethereumSender: ethereum sender,
    * @param _chain33Receiver: receiver on chain33
    * @param _tokenAddress: token address
    * @param _symbol: token symbol
    * @param _amount: amount
    * @param _claimID: claim id
    * @param _message: message for verifying
    * @param _signature: need to recover sender
    */
    function newOracleClaim(
        ClaimType _claimType,
        bytes memory _ethereumSender,
        address payable _chain33Receiver,
        address _tokenAddress,
        string memory _symbol,
        uint256 _amount,
        bytes32 _claimID,
        bytes memory _signature
    )
        public
        onlyValidator
        isValidClaimType(_claimType)
    {
        address validatorAddress = msg.sender;

        // Validate the msg.sender's signature
        require(
            validatorAddress == valset.recover(
                _claimID,
                _signature
            ),
            "Invalid _claimID signature."
        );

        // Confirm that this address has not already made an oracle claim on this _ClaimID
        require(
            !hasMadeClaim[_claimID][validatorAddress],
            "Cannot make duplicate oracle claims from the same address."
        );

        if (oracleClaimValidators[_claimID].length == 0) {
             ethereumBridge.setNewProphecyClaim(
                            _claimID,
                            uint8(_claimType),
                            _ethereumSender,
                            _chain33Receiver,
                            validatorAddress,
                            _tokenAddress,
                            _symbol,
                            _amount);
        }

        hasMadeClaim[_claimID][validatorAddress] = true;
        oracleClaimValidators[_claimID].push(validatorAddress);

        emit LogNewOracleClaim(
            _claimID,
            validatorAddress,
            _signature
        );

        (bool valid, uint256 weightedSignedPower, uint256 weightedTotalPower ) = getClaimThreshold(_claimID);
        if (true == valid)  {
            //if processed already,just emit an event
            if (ethereumBridge.isProphecyClaimActive(_claimID) == true) {
                completeClaim(_claimID);
            }
            emit LogProphecyProcessed(
                _claimID,
                weightedSignedPower,
                weightedTotalPower,
                msg.sender
            );
        }
    }

    /*
    * @dev: checkBridgeProphecy
    *       Operator accessor method which checks if a prophecy has passed
    *       the validity threshold, without actually completing the prophecy.
    */
    function checkBridgeProphecy(
        bytes32 _claimID
    )
        public
        view
        onlyOperator
        isPending(_claimID)
        returns(bool, uint256, uint256)
    {
        require(
            ethereumBridge.isProphecyClaimActive(
                _claimID
            ) == true,
            "Can only check active prophecies"
        );
        return getClaimThreshold(
            _claimID
        );
    }

    /*
    * @dev: getClaimThreshold
    *       Calculates the status of a claim. The claim is considered valid if the
    *       combined active signatory validator powers pass the validation threshold.
    *       The hardcoded threshold is (Combined signed power * 2) >= (Total power * 3).
    */
    function getClaimThreshold(
        bytes32 _claimID
    )
        internal
        view
        returns(bool, uint256, uint256)
    {
        uint256 signedPower = 0;
        uint256 totalPower = valset.totalPower();

        // Iterate over the signatory addresses
        for (uint256 i = 0; i < oracleClaimValidators[_claimID].length; i = i.add(1)) {
            address signer = oracleClaimValidators[_claimID][i];

                // Only add the power of active validators
                if(valset.isActiveValidator(signer)) {
                    signedPower = signedPower.add(
                        valset.getValidatorPower(
                            signer
                        )
                    );
                }
        }

        // Calculate if weighted signed power has reached threshold of weighted total power
        uint256 weightedSignedPower = signedPower.mul(3);
        uint256 weightedTotalPower = totalPower.mul(2);
        bool hasReachedThreshold = weightedSignedPower >= weightedTotalPower;

        return(
            hasReachedThreshold,
            weightedSignedPower,
            weightedTotalPower
        );
    }

    /*
    * @dev: completeClaim
    *       Completes a claim by completing the corresponding BridgeClaim
    *       on the EthereumBridge.
    */
    function completeClaim(
        bytes32 _claimID
    )
        internal
    {
        ethereumBridge.completeClaim(
            _claimID
        );
    }
}