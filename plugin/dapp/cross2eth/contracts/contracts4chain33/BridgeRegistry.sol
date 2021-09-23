pragma solidity ^0.5.0;

contract BridgeRegistry {

    address public ethereumBridge;
    address public bridgeBank;
    address public oracle;
    address public valset;
    uint256 public deployHeight;

    event LogContractsRegistered(
        address _ethereumBridge,
        address _bridgeBank,
        address _oracle,
        address _valset
    );
    
    constructor(
        address _ethereumBridge,
        address _bridgeBank,
        address _oracle,
        address _valset
    )
        public
    {
        ethereumBridge = _ethereumBridge;
        bridgeBank = _bridgeBank;
        oracle = _oracle;
        valset = _valset;
        deployHeight = block.number;

        emit LogContractsRegistered(
            ethereumBridge,
            bridgeBank,
            oracle,
            valset
        );
    }
}