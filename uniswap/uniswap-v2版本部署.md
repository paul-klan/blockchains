# uniswap v2版本部署

### 一、准备工作

合约间调用关系如下：

![sol-flow](/Users/yulei/Documents/solidity-flow.png)

##### 1.1 使用 remix 部署合约

打开 [remix](https://remix.ethereum.org/) 官网，推荐使用Chrome浏览器

##### 1.2 准备合约

1、Factory合约：

https://cn.etherscan.com/address/0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f#code

2、WETH合约：

https://cn.etherscan.com/address/0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2#code

3、Router合约02版本：

https://cn.etherscan.com/address/0x7a250d5630b4cf539739df2c5dacb4c659f2488d#code

4、multicall合约：

https://cn.etherscan.com/address/0x5e227ad1969ea493b43f840cff78d08a6fc17796#code

##### 1.3 准备前端代码仓库

### 二、部署合约

##### 2.1 部署WETH

- 编译设置
  - 新建文件 WETH.sol，将 `WETH合约`复制过来。
  - Optimization设置200， 使用默认环境编译
- 部署设置
  - CONTRACT 选择 WETH9，ENVIRONMENT 选择 Injected Web3
  - 连接 MetaMask 的 部署 网络进行部署。

##### 2.2 部署Multicall

- 编译设置
  - 新建文件 Multicall.sol，将 `Multicall合约`复制过来。
  - Optimization设置200， 使用默认环境编译
- 部署设置
  - CONTRACT 选择 Multicall，ENVIRONMENT 选择 Injected Web3
  - 连接 MetaMask 的 部署 网络进行部署。

##### 2.3 部署Factory

- 编译设置
  
  - 新建文件 UniswapV2Factory.sol，将 `Factory合约` 复制过来。
  
  - 需要修改，400行代码下面添加一行代码：
    
    `bytes32 public constant INIT_CODE_PAIR_HASH = keccak256(abi.encodePacked(type(UniswapV2Pair).creationCode));`
    
    ![factory](/Users/yulei/Documents/Factory1.png)
  
  - Optimization设置999999， 使用`istanbul` EVM版本编译

- 部署设置
  
  - 不设置收费的情况下，随便设置一个 _feeToSetter 地址，CONTRACT 选择UniswapV2Factory，ENVIRONMENT 选择 Injected Web3
  - 连接 MetaMask 的 部署 网络进行部署。

##### 2.4 部署Router

- 编译设置
  
  - 新建文件 UniswapV2Router02.sol，将 `Router合约02版本`复制过来。
  
  - 在remix中获取 UniswapV2Factory.sol 中 CONTRACT 为 UniswapV2Pair 时 Bytecode 的object对象的hash值（可通过查看UniswapV2Factory.sol中的init_code 哈希），替换 UniswapV2Router02.sol 中的 initCode 码，700行位置，可通过 `// init code hash` 查找，这样做是因为Router需要通过这个hash找到Pair的地址
    
    ![aa](/Users/yulei/Documents/router1.png)
  
  - Optimization设置999999， 使用`istanbul` EVM版本编译。

- 部署设置
  
  - 填入上面的 Factory 和 WETH 合约地址之后，CONTRACT 选择 UniswapV2Router02，ENVIRONMENT 选择 Injected Web3
  - 连接 MetaMask 的 部署 网络进行部署。

### 三、web调用
