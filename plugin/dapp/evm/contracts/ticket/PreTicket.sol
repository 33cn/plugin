//DO NOT EDIT.
pragma solidity ^0.8.0;

library PreTicket {
    /*
    * 自定义预编译合约地址,PRECOMPILE 不可修改
    */
    address constant private PRECOMPILE = address(0x0000000000000000000000000000000000200002);
    //BindMiner 建立绑定挖矿关系
    function createBindMiner(address origin, address bind, uint256 amount)  internal {
        (bool success, bytes memory returnData) = PRECOMPILE.call(
            abi.encodeWithSignature("bindMiner(address,address,uint256)", origin, bind, amount)
        );
        assembly {
            if eq(success, 0) {
                revert(add(returnData, 0x20), returndatasize())
            }
        }
    }
    //把amount 数量的币转账到ticket 执行器地址下
    function transferToTicketExec(address from,uint256 amount)internal{
        (bool success, bytes memory returnData) = PRECOMPILE.call(
            abi.encodeWithSignature("transferToTicketExec(address,uint256)", from, amount)
        );
        assembly {
            if eq(success, 0) {
                revert(add(returnData, 0x20), returndatasize())
            }
        }
    }


    function getTicketCount()internal view returns (uint256){
        (bool success, bytes memory returnData) = PRECOMPILE.staticcall(
            abi.encodeWithSignature("getTicketCount()")
        );
        assembly {
            if eq(success, 0) {
                revert(add(returnData, 0x20), returndatasize())
            }
        }

        return abi.decode(returnData, (uint256));
    }

    //close 挖矿，需要在open miner 48小时之后，先close ticket,然后解除绑定
    function closeTicketMiner(address bind,address owner) internal {
        (bool success, bytes memory returnData) = PRECOMPILE.call(
            abi.encodeWithSignature("closeTicketMiner")
        );
        assembly {
            if eq(success, 0) {
                revert(add(returnData, 0x20), returndatasize())
            }
        }
    }


}