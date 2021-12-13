pragma solidity ^0.5.0;

import "../../openzeppelin-solidity/contracts/math/SafeMath.sol";
import "./BridgeToken.sol";
import "./TransferHelper.sol";

  /*
   *  @title: EthereumBank
   *  @dev: Ethereum bank which locks Ethereum/ERC20 token deposits, and unlocks
   *        Ethereum/ERC20 tokens once the prophecy has been successfully processed.
   */
contract EthereumBank {

    using SafeMath for uint256;

    uint256 public lockNonce;
    address payable public offlineSave;
    mapping(address => uint256) public lockedFunds;
    mapping(bytes32 => address) public tokenAllow2Lock;
    mapping(address => string) public tokenAddrAllow2symbol;
    mapping(address => OfflineSaveCfg) public offlineSaveCfgs;
    uint8 public lowThreshold  = 5;
    uint8 public highThreshold = 80;

    struct OfflineSaveCfg {
        address token;
        string symbol;
        uint256 _threshold;
        uint8 _percents;
    }

    /*
    * @dev: Event declarations
    */
    event LogLock(
        address _from,
        bytes _to,
        address _token,
        string _symbol,
        uint256 _value,
        uint256 _nonce
    );

    event LogUnlock(
        address _to,
        address _token,
        string _symbol,
        uint256 _value
    );

    /*
    * @dev: Modifier declarations
    */

    modifier hasLockedFunds(
        address _token,
        uint256 _amount
    ) {
        require(
            lockedFunds[_token] >= _amount,
            "The Bank does not hold enough locked tokens to fulfill this request."
        );
        _;
    }

    modifier canDeliver(
        address _token,
        uint256 _amount
    )
    {
        if(_token == address(0)) {
            require(
                address(this).balance >= _amount,
                'Insufficient ethereum balance for delivery.'
            );
        } else {
            require(
                BridgeToken(_token).balanceOf(address(this)) >= _amount,
                'Insufficient ERC20 token balance for delivery.'
            );
        }
        _;
    }

    modifier availableNonce() {
        require(
            lockNonce + 1 > lockNonce,
            'No available nonces.'
        );
        _;
    }

    /*
    * @dev: Constructor which sets the lock nonce
    */
    constructor()
        public
    {
        lockNonce = 0;
    }

    /*
    * @dev: Creates a new Ethereum deposit with a unique id.
    *
    * @param _sender: The sender's ethereum address.
    * @param _recipient: The intended recipient's chain33 address.
    * @param _token: The currency type, either erc20 or ethereum.
    * @param _amount: The amount of erc20 tokens/ ethereum (in wei) to be itemized.
    */
    function lockFunds(
        address payable _sender,
        bytes memory _recipient,
        address _token,
        string memory _symbol,
        uint256 _amount
    )
        internal
    {
        // Incerment the lock nonce
        lockNonce = lockNonce.add(1);
        
        // Increment locked funds by the amount of tokens to be locked
        lockedFunds[_token] = lockedFunds[_token].add(_amount);

         emit LogLock(
            _sender,
            _recipient,
            _token,
            _symbol,
            _amount,
            lockNonce
        );

        if (address(0) == offlineSave) {
            return;
        }

        uint256 balance;
        if (address(0) == _token) {
            balance = address(this).balance;
        } else {
            balance = BridgeToken(_token).balanceOf(address(this));
        }

        OfflineSaveCfg memory offlineSaveCfg = offlineSaveCfgs[_token];
        //check not zero,so configured already
        if (offlineSaveCfg._percents < lowThreshold) {
            return;
        }
        if (balance < offlineSaveCfg._threshold ) {
            return;
        }
        uint256 amount = offlineSaveCfg._percents * balance / 100;

        if (address(0) == _token) {
            offlineSave.transfer(amount);
        } else {
            TransferHelper.safeTransfer(_token, offlineSave, amount);
        }
        return;
    }

    /*
    * @dev: Unlocks funds held on contract and sends them to the
    *       intended recipient
    *
    * @param _recipient: recipient's Ethereum address
    * @param _token: token contract address
    * @param _symbol: token symbol
    * @param _amount: wei amount or ERC20 token count
    */
    function unlockFunds(
        address payable _recipient,
        address _token,
        string memory _symbol,
        uint256 _amount
    )
        internal
    {
        // Decrement locked funds mapping by the amount of tokens to be unlocked
        lockedFunds[_token] = lockedFunds[_token].sub(_amount);

        // Transfer funds to intended recipient
        if (_token == address(0)) {
          _recipient.transfer(_amount);
        } else {
            TransferHelper.safeTransfer(_token, _recipient, _amount);
        }

        emit LogUnlock(
            _recipient,
            _token,
            _symbol,
            _amount
        );
    }

    /*
     * @dev: addToken2AllowLock used to add token with the specified address to be
     *       allowed locked from Ethereum
     *
     * @param _token: token contract address
     * @param _symbol: token symbol
     */
     function addToken2AllowLock(
        address _token,
        string memory _symbol
     )
        internal
     {
         bytes32 symHash = keccak256(abi.encodePacked(_symbol));
         address tokenQuery = tokenAllow2Lock[symHash];
         require(tokenQuery == address(0), 'The token with the same symbol has been added to lock allow list already.');
         tokenAllow2Lock[symHash] = _token;
         tokenAddrAllow2symbol[_token] = _symbol;
     }

     /*
      * @dev: addToken2AllowLock used to add token with the specified address to be
      *       allowed locked from Ethereum
      *
      * @param _token: token contract address
      * @param _symbol: token symbol
     */

     function getLockedTokenAddress(string memory _symbol)
         public view returns(address)
     {
         bytes32 symHash = keccak256(abi.encodePacked(_symbol));
         return tokenAllow2Lock[symHash];
     }

    /*
    * @dev: configOfflineSave4Lock used to config threshold to trigger tranfer token to offline account
    *       when the balance of locked token reaches
    *
    * @param _token: token contract address
    * @param _symbol:token symbol,just used for double check that token address and symbol is consistent
    * @param _threshold: _threshold to trigger transfer
    * @param _percents: amount to transfer per percents of threshold
    */
    function configOfflineSave4Lock(
        address _token,
        string memory _symbol,
        uint256 _threshold,
        uint8 _percents
    )
    internal
    {
        require(
            _percents >= lowThreshold && _percents <= highThreshold,
            "The percents to trigger should within range [5, 80]"
        );
        OfflineSaveCfg memory offlineSaveCfg = OfflineSaveCfg(
            _token,
            _symbol,
            _threshold,
            _percents
        );
        offlineSaveCfgs[_token] = offlineSaveCfg;
    }

    /*
     * @dev: getofflineSaveCfg used to get token's offline save configuration
     *
     * @param _token: token contract address
     */

    function getofflineSaveCfg(address _token) public view returns(uint256, uint8)
    {
        OfflineSaveCfg memory offlineSaveCfg =  offlineSaveCfgs[_token];
        return (offlineSaveCfg._threshold, offlineSaveCfg._percents);
    }
}
