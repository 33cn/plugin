pragma solidity ^0.5.0;

contract BridgeRegistry {

    address public goAssetBridge;
    address public bridgeBank;
    address public oracle;
    address public valset;
    uint256 public deployHeight;

    event LogContractsRegistered(
        address _goAssetBridge,
        address _bridgeBank,
        address _oracle,
        address _valset
    );
    
    constructor(
        address _goAssetBridge,
        address _bridgeBank,
        address _oracle,
        address _valset
    )
        public
    {
        goAssetBridge = _goAssetBridge;
        bridgeBank = _bridgeBank;
        oracle = _oracle;
        valset = _valset;
        deployHeight = block.number;

        emit LogContractsRegistered(
            goAssetBridge,
            bridgeBank,
            oracle,
            valset
        );
    }
}