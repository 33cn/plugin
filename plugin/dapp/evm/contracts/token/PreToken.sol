//DO NOT EDIT.
pragma solidity ^0.8.0;

library PreToken {
    /*
    * 自定义预编译合约地址,PRECOMPILE 不可修改
    */
    address constant private PRECOMPILE = address(0x0000000000000000000000000000000000200001);
    function decimals() internal view returns (uint8) {
        (bool success, bytes memory returnData) = PRECOMPILE.staticcall(abi.encodeWithSignature("decimals()"));
        assembly {
            if eq(success, 0) {
                revert(add(returnData, 0x20), returndatasize())
            }
        }

        return abi.decode(returnData, (uint8));
    }

    function totalSupply()internal view returns(uint256){
        (bool success, bytes memory returnData) = PRECOMPILE.staticcall(abi.encodeWithSignature("totalSupply()"));
        assembly {
            if eq(success, 0) {
                revert(add(returnData, 0x20), returndatasize())
            }
        }

        return abi.decode(returnData, (uint256));
    }

    function balanceOf(address account) internal view returns (uint256) {
        (bool success, bytes memory returnData) = PRECOMPILE.staticcall(
            abi.encodeWithSignature("balanceOf(address)", account)
        );
        assembly {
            if eq(success, 0) {
                revert(add(returnData, 0x20), returndatasize())
            }
        }

        return abi.decode(returnData, (uint256));
    }

    function transfer(address sender, address recipient, uint256 amount)  internal {
        (bool success, bytes memory returnData) = PRECOMPILE.call(
            abi.encodeWithSignature("transfer(address,address,uint256)", sender, recipient, amount)
        );
        assembly {
            if eq(success, 0) {
                revert(add(returnData, 0x20), returndatasize())
            }
        }
    }
}