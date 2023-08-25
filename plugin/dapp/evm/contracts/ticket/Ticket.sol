//DO NOT EDIT.
pragma solidity ^0.8.0;

import "./PreTicket.sol";

contract Ticket {
    //构造函数
    constructor(){

    }
    //事件
    event CreateBindMiner(address origin,address bind,uint256 amount);
    event TransferToTicketExec(address from,uint256 amount);
    event CloseBindMiner(address origin,address bind);
   /*
     * @dev See {ticket.md}.
     *
     * Requirements:
     * - `origin` client address
     * - `bind`   miner address.
     * - `amount` pay for ticket.
   */

    function createBindMiner(address bind,uint256 amount) public  returns (bool){
        require(bind != address(0), "Ticket: createBindMiner from the  zero address");
        require(amount != 0, "Ticket: createBindMiner amount is zero");
        address owner = msg.sender;
        //调用预编译合约地址,
        PreTicket.createBindMiner(owner, bind, amount);
        //触发CreateBindMiner事件
        emit CreateBindMiner(owner, bind, amount);
        return true;
    }

    /*
    * @dev See {ticket.md}.
    *
    * Requirements:
    * - `from` client address
    * - `amount` pay for ticket to ticket address.
  */

    function transferToTickeExec(uint256 amount)public  returns (bool){
        require(amount != 0, "Ticket: transferToTicketExec amount is zero");
        address owner = msg.sender;
        PreTicket.transferToTicketExec(owner,amount);
        emit TransferToTicketExec(owner, amount);
        return true;
    }

    function getTicketCount()public  returns (uint256){
        return PreTicket.getTicketCount();
    }

    function closeBindMiner(address bind)public returns(bool){
        require(bind != address(0), "Ticket: createBindMiner from the  zero address");
        PreTicket.closeTicketMiner(bind,msg.sender);
        emit CloseBindMiner(msg.sender,bind);
        return true;
    }
}