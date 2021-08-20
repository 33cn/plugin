pragma solidity ^0.5.0;

contract BridgeRegistry {

    address public chain33Bridge;
    address public bridgeBank;
    address public oracle;
    address public valset;
    uint256 public deployHeight;

    event LogContractsRegistered(
        address _chain33Bridge,
        address _bridgeBank,
        address _oracle,
        address _valset
    );
    
    constructor(
        address _chain33Bridge,
        address _bridgeBank,
        address _oracle,
        address _valset
    )
        public
    {
        chain33Bridge = _chain33Bridge;
        bridgeBank = _bridgeBank;
        oracle = _oracle;
        valset = _valset;
        deployHeight = block.number;

        emit LogContractsRegistered(
            chain33Bridge,
            bridgeBank,
            oracle,
            valset
        );
    }
}