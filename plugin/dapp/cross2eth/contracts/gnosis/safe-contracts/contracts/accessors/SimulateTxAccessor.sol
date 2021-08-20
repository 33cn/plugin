pragma solidity >=0.7.0 <0.9.0;

import "../base/Executor.sol";

/// @title Simulate Transaction Accessor - can be used with StorageAccessible to simulate Safe transactions
/// @author Richard Meissner - <richard@gnosis.pm>
contract SimulateTxAccessor is Executor {

    bytes32 constant private GUARD_VALUE = keccak256("simulate_tx_accessor.guard.bytes32");
    bytes32 guard;

    constructor() public {
        guard = GUARD_VALUE;
    }

    modifier onlyDelegateCall() {
        require(guard != GUARD_VALUE, "SimulateTxAccessor should only be called via delegatecall");
        _;
    }

    function simulate(address to, uint256 value, bytes calldata data, Enum.Operation operation)
        external
        onlyDelegateCall()
        returns(uint256 estimate, bool success, bytes memory returnData)
    {
        uint256 startGas = gasleft();
        success = execute(to, value, data, operation, gasleft());
        estimate = startGas - gasleft();
        // solium-disable-next-line security/no-inline-assembly
        assembly {
            // Load free memory location
            let ptr := mload(0x40)
            // We allocate memory for the return data by setting the free memory location to
            // current free memory location + data size + 32 bytes for data size value
            mstore(0x40, add(ptr, add(returndatasize(), 0x20)))
            // Store the size
            mstore(ptr, returndatasize())
            // Store the data
            returndatacopy(add(ptr, 0x20), 0, returndatasize())
            // Point the return data to the correct memory location
            returnData := ptr
        }
    }
}