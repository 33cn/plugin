pragma solidity ^0.5.0;

import "../../openzeppelin-solidity/contracts/token/ERC20/ERC20Mintable.sol";
import "../../openzeppelin-solidity/contracts/token/ERC20/ERC20Burnable.sol";
import "../../openzeppelin-solidity/contracts/token/ERC20/ERC20Detailed.sol";

/**
 * @title BridgeToken
 * @dev Mintable, ERC20 compatible BankToken for use by BridgeBank
 **/

contract BridgeToken is ERC20Mintable, ERC20Burnable, ERC20Detailed {

    constructor(
        string memory _symbol
    )
        public
        ERC20Detailed(
            _symbol,
            _symbol,
            8
        )
    {
        // Intentionally left blank
    }
}