pragma solidity ^0.5.0;

// helper methods for interacting with ERC20 tokens and sending ETH that do not consistently return true/false
library TransferHelper {
    function safeApprove(address token, address to, uint value) internal {
        // bytes4(keccak256(bytes('approve(address,uint256)')));
        (bool success, bytes memory data) = token.call(abi.encodeWithSelector(0x095ea7b3, to, value));
        require(success && (data.length == 0 || abi.decode(data, (bool))), 'TransferHelper: APPROVE_FAILED');
    }

    function safeTransfer(address token, address to, uint value) internal {
        // bytes4(keccak256(bytes('transfer(address,uint256)')));
        (bool success, bytes memory data) = token.call(abi.encodeWithSelector(0xa9059cbb, to, value));
        require(success && (data.length == 0 || abi.decode(data, (bool))), 'TransferHelper: TRANSFER_FAILED');
    }

    function safeTransferFrom(address token, address from, address to, uint value) internal {
        // bytes4(keccak256(bytes('transferFrom(address,address,uint256)')));
        (bool success, bytes memory data) = token.call(abi.encodeWithSelector(0x23b872dd, from, to, value));
        require(success && (data.length == 0 || abi.decode(data, (bool))), 'TransferHelper: TRANSFER_FROM_FAILED');
    }

//    function safeTransferETH(address to, uint value) internal {
//        (bool success,) = to.call{value:value}(new bytes(0));
//        require(success, 'TransferHelper: ETH_TRANSFER_FAILED');
//    }
}

//var ERC20FuncSigs = map[string]string{
//"dd62ed3e": "allowance(address,address)",
//"095ea7b3": "approve(address,uint256)",
//"70a08231": "balanceOf(address)",
//"313ce567": "decimals()",
//"a457c2d7": "decreaseAllowance(address,uint256)",
//"39509351": "increaseAllowance(address,uint256)",
//"06fdde03": "name()",
//"95d89b41": "symbol()",
//"18160ddd": "totalSupply()",
//"a9059cbb": "transfer(address,uint256)",
//"23b872dd": "transferFrom(address,address,uint256)",
//}
