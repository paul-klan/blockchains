# # 剖析DeFi交易产品之Uniswap（Keegan小钢）

原文：

[剖析DeFi交易产品之Uniswap：V2上篇 | 登链社区 | 深入浅出区块链技术](https://learnblockchain.cn/article/2824)

 [剖析DeFi交易产品之Uniswap：V2中篇 | 登链社区 | 深入浅出区块链技术](https://learnblockchain.cn/article/3047)

[剖析DeFi交易产品之Uniswap：V2下篇 | 登链社区 | 深入浅出区块链技术](https://learnblockchain.cn/article/3100)



从使用用途角度来看，合约调用顺序如下面

* **管理流动性时**
  
  ## ![](/Users/yulei/Documents/liq.jpeg)

* **交易时**

![swap](/Users/yulei/Documents/swap.jpeg)

合约间的相互调用关系如下：

![sodility-flow](/Users/yulei/Documents/solidity-flow.png)

### 一、uniswap-v2-core

core 核心主要有三个合约文件：

- **UniswapV2Factory.sol**：工厂合约
- **UniswapV2Pair.sol**：配对合约
- **UniswapV2ERC20.sol**：LP Token 合约

**配对合约**管理着流动性资金池，不同币对有着不同的配对合约实例，比如 USDT-WETH 这一个币对，就对应一个配对合约实例，DAI-WETH 又对应另一个配对合约实例。

**LP Token** 则是用户往资金池里注入流动性的一种凭证，也称为**流动性代币**，本质上和 **Compound 的 cToken** 类似。当用户往某个币对的配对合约里转入两种币，即添加流动性，就可以得到配对合约返回的 LP Token，享受手续费分成收益。

每个配对合约都有对应的一种 LP Token 与之绑定。其实，**UniswapV2Pair** 继承了 **UniswapV2ERC20**，所以配对合约本身其实也是 LP Token 合约。

**工厂合约**则是用来部署配对合约的，通过工厂合约的 **createPair()** 函数来创建新的配对合约实例。



#### 1.1 工厂合约

工厂合约最核心的函数就是 **createPair()** ，其实现代码如下：

![](https://pic2.zhimg.com/80/v2-1f06d3090c5ef396f81e47b784423119_1440w.jpg)

里面创建合约采用了 **create2**，这是一个汇编 **opcode**，这是我要重点讲解的部分。

很多小伙伴应该都知道，一般创建新合约可以使用 **new** 关键字，比如，创建一个新配对合约，也可以这么写：

```
UniswapV2Pair newPair = new UniswapV2Pair();
```

那为什么不使用 new 的方式，而是调用 create2 操作码来新建合约呢？使用 create2 最大的好处其实在于：**可以在部署智能合约前预先计算出合约的部署地址**。最关键的就是以下这几行代码：

```assembly
bytes memory bytecode = type(UniswapV2Pair).creationCode;
bytes32 salt = keccak256(abi.encodePacked(token0, token1));
assembly {
  pair := create2(0, add(bytecode, 32), mload(bytecode), salt)
}
```

第一行获取 **UniswapV2Pair** 合约代码的创建字节码 **creationCode**，结果值一般是这样：

```
0x0cf061edb29fff92bda250b607ac9973edf2282cff7477decd42a678e4f9b868
```

类似的，其实还有运行时的字节码 **runtimeCode**，但这里没有用到。

这个创建字节码其实会在 **periphery** 项目中的 **UniswapV2Library** 库中用到，是被硬编码设置的值。所以为了方便，可以在工厂合约中添加一行代码保存这个创建字节码：

```assembly
bytes32 public constant INIT_CODE_PAIR_HASH = keccak256(abi.encodePacked(type(UniswapV2Pair).creationCode));
```

回到上面代码，第二行根据两个代币地址计算出一个盐值，对于任意币对，计算出的盐值也是固定的，所以也可以线下计算出该币对的盐值。

接着就用 **assembly** 关键字包起一段**内嵌汇编代码**，里面调用 **create2** 操作码来创建新合约。因为 UniswapV2Pair 合约的创建字节码是固定的，两个币对的盐值也是固定的，所以最终计算出来的 pair 地址其实也是固定的。

除了 create2 创建新合约的这部分代码之外，其他的都很好理解，我就不展开说明了。

#### 1.2  UniswapV2ERC20合约

配对合约继承了 *UniswapV2ERC20* 合约，我们先来看看 *UniswapV2ERC20* 合约的实现，这个比较简单。

*UniswapV2ERC20* 是**流动性代币**合约，也称为 **LP Token**，但代币实际名称为 **Uniswap V2**，简称为 **UNI-V2**，都是直接在代码中定义好的：

```typescript
string public constant name = 'Uniswap V2';
string public constant symbol = 'UNI-V2';
```

而代币的总量 **totalSupply** 最初为 0，可通过调用 **_mint()** 函数铸造出来，还可通过调用 **_burn()** 进行销毁。这两个函数的代码实现非常简单，就是直接在 **totalSupply** 和指定账户的 **balance** 上进行加减，只是，两个函数都是 **internal** 的，所以无法外部调用，代码如下：

```assembly
function _mint(address to, uint value) internal {
  totalSupply = totalSupply.add(value);
  balanceOf[to] = balanceOf[to].add(value);
  emit Transfer(address(0), to, value);
}

function _burn(address from, uint value) internal {
  balanceOf[from] = balanceOf[from].sub(value);
  totalSupply = totalSupply.sub(value);
  emit Transfer(from, address(0), value);
}
```

另外，*UniswapV2ERC20* 还提供了一个 **permit()** 函数，它允许用户在链下签署授权（approve）的交易，生成任何人都可以使用并提交给区块链的签名。关于 permit 函数具体的作用和用法，网上已经有很多介绍文章，我这里就不展开了。

除此之后，剩下的都是符合 *ERC20* 标准的函数了。

#### 1.3  配对合约

前面说过，配对合约是由工厂合约创建的，我们从构造函数和初始化函数中就可以看出来：

```typescript
constructor() public {
    factory = msg.sender;
}

// called once by the factory at time of deployment
function initialize(address _token0, address _token1) external {
  require(msg.sender == factory, 'UniswapV2: FORBIDDEN'); // sufficient check
  token0 = _token0;
  token1 = _token1;
}
```

构造函数直接将 **msg.sender** 设为了 **factory** ，*factory* 就是工厂合约地址。初始化函数又 require 调用者需是工厂合约，而且工厂合约中只会初始化一次。

不过，不知道你有没有想到，为什么还要另外定义一个初始化函数，而不直接将 **_token0** 和 **_token1** 在构造函数中作为入参进行初始化呢？这是因为用 **create2** 创建合约的方式限制了构造函数不能有参数。

另外，配对合约中最核心的函数有三个：**mint()、burn()、swap()** 。分别是**添加流动性、移除流动性、兑换**三种操作的底层函数。

##### 1.3.1 mint() 函数

先来看看 *mint()* 函数，主要是通过同时注入两种代币资产来获取流动性代币：

![](https://pic3.zhimg.com/80/v2-bc297119bd82fb4ec05590add31b9892_1440w.jpg)

既然这是一个添加流动性的底层函数，那参数里为什么没有两个代币投入的数量呢？这可能是大部分人会想到的第一个问题。其实，调用该函数之前，**路由合约**已经完成了将用户的代币数量划转到该配对合约的操作。因此，你看前五行代码，通过获取两个币的当前余额 *balance0* 和 *balance1*，再分别减去 *_reserve0* 和 *_reserve1*，即池子里两个代币原有的数量，就计算得出了两个代币的投入数量 *amount0* 和 *amount1*。另外，还给该函数添加了 **lock** 的修饰器，这是一个防止重入的修饰器，保证了每次添加流动性时不会有多个用户同时往配对合约里转账，不然就没法计算用户的 *amount0* 和 *amount1* 了。

第 6 行代码是计算协议费用的。在工厂合约中有一个 **feeTo** 的地址，如果设置了该地址不为零地址，就表示添加和移除流动性时会收取协议费用，但 Uniswap 一直到现在都没有设置该地址。

接着从第 7 行到第 15 行代码则是计算用户能得到多少流动性代币了。当 **totalSupply** 为 0 时则是最初的流动性，计算公式为：

```mathematica
liquidity = √(amount0*amount1) - MINIMUM_LIQUIDITY
```

即两个代币投入的数量相乘后求平方根，结果再减去最小流动性。最小流动性为 1000，该最小流动性会永久锁在零地址。这么做，主要还是为了安全，具体原因可以查看白皮书和官方文档的说明。

如果不是提供最初流动性的话，那流动性则是取以下两个值中较小的那个：

```mathematica
liquidity1 = amount0 * totalSupply / reserve0
liquidity2 = amount1 * totalSupply / reserve1
```

计算出用户该得的流动性 **liquidity** 之后，就会调用前面说的 *_mint()* 函数铸造出 *liquidity* 数量的 *LP Token* 并给到用户。

接着就会调用 **_update()** 函数，该函数主要做两个事情，一是更新 *reserve0* 和 *reserve1*，二是累加计算 *price0CumulativeLast* 和 *price1CumulativeLast*，这两个价格是用来计算 TWAP 的，后面再讲。

倒数第 2 行则是判断如果协议费用开启的话，更新 **kLast** 值，即 *reserve0* 和 *reserve1* 的乘积值，该值其实只在计算协议费用时用到。

最后一行就是触发一个 *Mint()* 事件的发出。

##### 1.3.2 burn() 函数

接着就来看看 *burn()* 函数了，这是**移除流动性**的底层函数：

![](https://pic4.zhimg.com/80/v2-bfa09c8fc17215af0979231e3e9eb73f_1440w.jpg)

该函数主要就是销毁掉流动性代币并提取相应的两种代币资产给到用户。

这里面第一个不太好理解的就是第 6 行代码，获取当前合约地址的流动性代币余额。正常情况下，配对合约里是不会有流动性代币的，因为所有流动性代币都是给到了流动性提供者的。而这里有值，其实是因为**路由合约**会先把用户的流动性代币划转到该配对合约里。

第 7 行代码计算协议费用和 mint() 函数一样的。

接着就是计算两个代币分别可以提取的数量了，计算公式也很简单：

```mathematica
amount = liquidity / totalSupply * balance
提取数量 = 用户流动性 / 总流动性 * 代币总余额
```

我调整了下计算顺序，这样就能更好理解了。用户流动性除以总流动性就得出了用户在整个流动性池子里的占比是多少，再乘以代币总余额就得出用户应该分得多少代币了。举例：用户的 liquidity 为 1000，totalSupply 有 10000，即是说用户的流动性占比为 10%，那假如池子里现在代币总额有 2000 枚，那用户就可分得这 2000 枚的 10% 即 200 枚。

后面的逻辑就是调用 **_burn()** 销毁掉流动性代币，且将两个代币资产计算所得数量划转给到用户，最后更新两个代币的 reserve。

最后两行代码也和 mint() 函数一样，就不赘述了。

##### 1.3.3 swap() 函数

**swap的总体流程：**

![](https://pic3.zhimg.com/80/v2-1f0ad0ae99ce21e16c776df8c40060ae_1440w.jpg)

swap() 就是做兑换交易的底层函数了，来看看代码：

![](https://pic3.zhimg.com/80/v2-dec11665d2f1940cfc817e5d6315076a_1440w.jpg)

该函数有 4 个入参，*amount0Out* 和 *amount1Out* 表示兑换结果要转出的 token0 和 token1 的数量，这两个值通常情况下是一个为 0，一个不为 0，但使用闪电交易时可能两个都不为 0。*to* 参数则是接收者地址，最后的 *data* 参数是执行回调时的传递数据，通过路由合约兑换的话，该值为 0。

前 3 行代码很好理解，第一步先校验兑换结果的数量是否有一个大于 0，然后读取出两个代币的 *reserve*，之后再校验兑换数量是否小于 *reserve*。

从第 6 行开始，到第 15 行结束，用了一对大括号，这主要是为了限制 *_token{0,1}* 这两个临时变量的作用域，防止堆栈太深导致错误。

接着，看看第 10 和 11 行，就开始将代币划转到接收者地址了。看到这里，有些小伙伴可能会产生疑问：这是个 *external* 函数，任何用户都可以自行调用的，没有校验就直接划转了，那不是谁都可以随便提币了？其实，在后面是有校验的，我们往下看就知道了。

第 12 行，如果 *data* 参数长度大于 0，则将 *to* 地址转为 *IUniswapV2Callee* 并调用其 *uniswapV2Call()* 函数，这其实就是一个回调函数，*to* 地址需要实现该接口。

第 13 和 14 行，获取两个代币当前的余额 *balance{0,1}* ，而这个余额是扣减了转出代币后的余额。

第 16 和 17 行则是计算出实际转入的代币数量了。实际转入的数量其实也通常是一个为 0，一个不为 0 的。要理解计算公式的原理，我举一个实例来说明。

假设转入的是 token0，转出的是 token1，转入数量为 100，转出数量为 200。那么，下面几个值将如下：

```mathematica
amount0In = 100
amount1In = 0
amount0Out = 0
amount1Out = 200
```

而 *reserve0* 和 *reserve1* 假设分别为 1000 和 2000，没进行兑换交易之前，*balance{0,1}* 和 *reserve{0,1}* 是相等的。而完成了代币的转入和转出之后，其实，*balance0* 就变成了 1000 + 100 - 0 = 1100，*balance1* 变成了 2000 + 0 - 200 = 1800。整理成公式则如下：

```mathematica
balance0 = reserve0 + amount0In - amout0Out
balance1 = reserve1 + amount1In - amout1Out
```

反推一下就得到：

```mathematica
amountIn = balance - (reserve - amountOut)
```

这下就明白代码里计算 *amountIn* 背后的逻辑了吧。

之后的代码则是进行扣减交易手续费后的恒定乘积校验，使用以下公式：

![a](https://pic1.zhimg.com/80/v2-19a98bf43932e38f0940c0974d53af38_1440w.png)

其中，*0.003* 是交易手续费率，*X0* 和 *Y0* 就是 *reserve0* 和 *reserve1*，*X1* 和 *Y1* 则是 *balance0* 和 *balance1*，*Xin* 和 *Yin* 则对应于 *amount0In* 和 *amount1In*。该公式成立就说明在进行这个底层的兑换之前的确已经收过交易手续费了。

### 二、 uniswap-v2-periphery

*periphery* 项目的结构很简单，如下：

- **UniswapV2Migrator.sol**：迁移合约，从 V1 迁移到 V2 的合约
- **UniswapV2Router01.sol**：路由合约 01 版本
- **UniswapV2Router02.sol**：路由合约 02 版本，相比 01 版本主要增加了几个支持交税费用的函数
- **interfaces**：接口都统一放在该目录下
- **libraries**：存放用到的几个库文件
- **test**：里面有几个测试用的合约
- **examples**：一些很有用的示例合约，包括 TWAP、闪电兑换等

当然，我们没必要每个合约都讲，主要讲解最核心的 **UniswapV2Router02.sol**，即路由合约。

#### 2.1 UniswapV2Library

讲路由合约之前，我想先聊聊 *UniswapV2Library* 这个库，路由合约很多函数的实现逻辑都用到了这个库提供的函数。 *UniswapV2Library* 主要提供了以下这些函数：

- **sortTokens**：对两个 token 进行排序
- **pairFor**：计算出两个 token 的 pair 合约地址
- **getReserves**：获取两个 token 在池子里里的储备量
- **quote**：根据给定的两个 token 的储备量和其中一个 token 数量，计算得到另一个 token 等值的数值
- **getAmountOut**：根据给定的两个 token 的储备量和输入的 token 数量，计算得到输出的 token 数量，该计算会扣减掉 0.3% 的手续费
- **getAmountIn**：根据给定的两个 token 的储备量和输出的 token 数量，计算得到输入的 token 数量，该计算会扣减掉 0.3% 的手续费
- **getAmountsOut**：根据兑换路径和输入数量，计算得到兑换路径中每个交易对的输出数量
- **getAmountsIn**：根据兑换路径和输出数量，计算得到兑换路径中每个交易对的输入数量

其中，第一个关键函数就是 *pairFor*，用来计算得到两个 token 的配对合约地址，其代码实现是这样的：

![image20210920140902958.png](https://img.learnblockchain.cn/attachments/2021/09/qsAqQFxt614ddaffe6aab.png)

可以看到，有个「**init code hash**」是硬编码的。该值其实是 **UniswapV2Pair** 合约的 *creationCode* 的哈希值。在「[上篇](https://mp.weixin.qq.com/s?__biz=MzA5OTI1NDE0Mw==&mid=2652494337&idx=1&sn=8a007959e5535b2603a6a0e1096be702&chksm=8b685011bc1fd907f84fbed1969c3240d66d70c7c724295ee3c9c2008271e8b788c302852406&token=276562139&lang=zh_CN#rd)」我们有提到，可以在 **UniswapV2Factory** 合约中添加以下常量获取到该值：

```solidity
bytes32 public constant INIT_CODE_PAIR_HASH = keccak256(abi.encodePacked(type(UniswapV2Pair).creationCode));
```

另外，*INIT_CODE_PAIR_HASH* 的值是带有 0x 开头的。而以上硬编码的 init code hash 前面已经加了 hex 关键字，所以单引号里的哈希值就不再需要 0x 开头。

接着，来看看 *getAmountOut* 的实现：

![image20210921145914470.png](https://img.learnblockchain.cn/attachments/2021/09/41lW7aYS614ddb1cbe3c7.png)

根据 AMM 的原理，恒定乘积公式「x * y = K」，兑换前后 K 值不变。因此，在不考虑交易手续费的情况下，以下公式会成立：

```mathematica
reserveIn * reserveOut = (reserveIn + amountIn) * (reserveOut - amountOut)
```

将公式右边的表达式展开，并推导下，就变成了：

```mathematica
reserveIn * reserveOut = reserveIn * reserveOut + amountIn * reserveOut - (reserveIn + amountIn) * amountOut
->
amountIn * reserveOut = (reserveIn + amountIn) * amountOut
->
amountOut = amountIn * reserveOut / (reserveIn + amountIn)
```

而实际上交易时，还需要扣减千分之三的交易手续费，所以实际上：

```mathematica
amountIn = amountIn * 997 / 1000
```

代入上面的公式后，最终结果就变成了：

```mathematic
amountOut = (amountIn * 997 / 1000) * reserverOut / (reserveIn + amountIn * 997 / 1000)
->
amountOut = amountIn * 997 * reserveOut / 1000 * (reserveIn + amountIn * 997 / 1000)
->
amountOut = amountIn * 997 * reserveOut / (reserveIn * 1000 + amountIn * 997)
```

这即是最后代码实现中的计算公式了。

*getAmountIn* 是类似的，就不展开说明了。

最后，再来看看 *getAmountsOut* 的代码实现：

![image20210921163841910.png](https://img.learnblockchain.cn/attachments/2021/09/CNkNZti3614ddb8ac6337.png)

该函数会计算 path 中每一个中间资产和最终资产的数量，比如 path 为 [A,B,C]，则会先将 A 兑换成 B，再将 B 兑换成 C。返回值则是一个数组，第一个元素是 A 的数量，即 amountIn，而第二个元素则是兑换到的代币 B 的数量，最后一个元素则是最终要兑换得到的代币 C 的数量。

从代码中还可看到，每一次兑换其实都调用了 *getAmountOut* 函数，这也意味着每一次中间兑换都会扣减千分之三的交易手续费。那如果兑换两次，实际支付假设为 1000，那最终实际兑换得到的价值只剩下：

```mathematica
1000 * (1 - 0.003) * (1 - 0.003) = 994.009
```

即实际支付的交易手续费将近千分之六了。兑换路径越长，实际扣减的交易手续费会更多，所以兑换路径一般不宜过长。

#### 2.2  UniswapV2Router02

*UniswapV2Router02* 路由合约是与用户进行交互的入口，主要提供了**添加流动性、移除流动性**和**兑换**的系列接口，并提供了几个查询接口。

##### 2.2.1 添加流动性接口

添加流动性，本质上就是支付两种代币，换回对应这两种代币的流动性代币 LP-Token。

添加流动性的接口有两个：

- **addLiquidity**：该接口支持添加两种 ERC20 代币作为流动性
- **addLiquidityETH**：与上一个接口不同，该接口提供的流动性资产，其中有一个是 ETH

我们先来看看第一个接口的实现代码：

![image20210921171836515.png](https://img.learnblockchain.cn/attachments/2021/09/pV12aaUN614ddbb782c06.png)

先介绍下该接口的几个入参。tokenA 和 tokenB 就是配对的两个代币，tokenADesired 和 tokenBDesired 是预期支付的两个代币的数量，amountAMin 和 amountBMin 则是用户可接受的最小成交数量，to 是接收流动性代币的地址，deadline 是该笔交易的有效时间，如果超过该时间还没得到交易处理就直接失效不进行交易了。

这几个参数，amountAMin 和 amountBMin 有必要再补充说明一下。该值一般是由前端根据预期值和滑点值计算得出的。比如，预期值 amountADesired 为 1000，设置的滑点为 0.5%，那就可以计算得出可接受的最小值 amountAMin 为 1000 * (1 - 0.5%) = 995。

再来看代码实现逻辑，第一步是先调用内部函数 _addLiquidity()。来看看该函数的实现代码：

![image20210921194054694.png](https://img.learnblockchain.cn/attachments/2021/09/Kn4aK2d2614ddbd09538a.png)

该函数的返回值 amountA 和 amountB 是最终需要支付的数量。

实现逻辑还是比较简单的。先通过工厂合约查一下这两个 token 的配对合约是否已经存在，如果不存在则先创建该配对合约。接着读取出两个 token 的储备量，如果储备量都为 0，那两个预期支付额就是成交量。否则，根据两个储备量和 tokenA 的预期支付额，计算出需要支付多少 tokenB，如果计算得出的结果值 amountBOptimal 不比 amountBDesired 大，且不会小于 amountBMin，就可将 amountADesired 和该 amountBOptimal 作为结果值返回。如果 amountBOptimal 大于 amountBDesired，则根据 amountBDesired 计算得出需要支付多少 tokenA，得到 amountAOptimal，只要 amountAOptimal 不大于 amountADesired 且不会小于 amountAMin，就可将 amountAOptimal 和 amountBDesired 作为结果值返回。

再回到 addLiquidity 函数的实现，计算得出两个 token 实际需要支付的数量之后，调用了 UniswapV2Library 的 pairFor 函数计算出配对合约地址，接着就往 pair 地址进行转账了。因为用了 transferFrom 的方式，所以用户调用该函数之前，其实是需要先授权给路由合约的。

最后再调用 pair 合约的 mint 接口就可以得到流动性代币 liquidity 了。

以上就是 addLiquidity 的基本逻辑，很简单，所以非常好理解。

而 addLiquidityETH 则支付的其中一个 token 则是 ETH，而不是 ERC20 代币。来看看其代码实现：

![image20210921212944219.png](https://img.learnblockchain.cn/attachments/2021/09/hS1DprXj614ddbf09dce7.png)

可看到，入参不再是两个 token 地址，而只有一个 token 地址，因为另一个是以太坊主币 ETH。预期支付的 ETH 金额也是直接从 msg.value 读取的，所以入参里也不需要 ETH 的 Desired 参数。但是会定义 amountETHMin 表示愿意接受成交的 ETH 最小额。

实现逻辑上，请注意，调用 _addLiquidity 时传入的第二个参数是 WETH。其实，addLiquidityETH 实际上也是将 ETH 转为 WETH 进行处理的。可以看到代码中还有这么一行：

```solidity
IWETH(WETH).deposit{value: amountETH}();
```

这就是将用户转入的 ETH 转成了 WETH。

而最后一行代码则会判断，如果一开始支付的 msg.value 大于实际需要支付的金额，多余的部分将返还给用户。

##### 2.2.2 移除流动性接口

移除流动性本质上就是用流动性代币兑换出配对的两个币。

移除流动性的接口有 6 个：

- **removeLiquidity**：和 addLiquidity 相对应，会换回两种 ERC20 代币
- **removeLiquidityETH**：和 addLiquidityETH 相对应，换回的其中一种是主币 ETH
- **removeLiquidityWithPermit**：也是换回两种 ERC20 代币，但用户会提供签名数据使用 permit 方式完成授权操作
- **removeLiquidityETHWithPermit**：也是使用 permit 完成授权操作，换回的其中一种是主币 ETH
- **removeLiquidityETHSupportingFeeOnTransferTokens**：名字真长，功能和 removeLiquidityETH 一样，不同的地方在于支持转账时支付费用
- **removeLiquidityETHWithPermitSupportingFeeOnTransferTokens**：功能和上一个函数一样，但支持使用链下签名的方式进行授权

removeLiquidity 是这些接口中最核心的一个，也是其它几个接口的元接口。来看看其代码实现？

![image20210921231526704.png](https://img.learnblockchain.cn/attachments/2021/09/rTVRm0Z4614ddc0f6c5eb.png)

代码逻辑很简单，就 7 行代码。第一行，先计算出 pair 合约地址；第二行，将流动性代币从用户划转到 pair 合约；第三行，执行 pair 合约的 burn 函数实现底层操作，返回了两个代币的数量；第四行对两个代币做下排序；第五行根据排序结果确定 amountA 和 amountB；最后两行检验是否大于滑点计算后的最小值。

removeLiquidityETH 也同样简单，其实现代码如下：

![image20210921233738867.png](https://img.learnblockchain.cn/attachments/2021/09/fls5SXP6614ddc26676d4.png)

因为流动性池子里实际存储的是 WETH，所以第一步调用 removeLiquidity 时第二个参数传的是 WETH。之后再调用 WETH 的 withdraw 函数将 WETH 转为 ETH，再将 ETH 转给用户。

removeLiquidityWithPermit 则是使用链下签名进行授权操作的，实现代码如下：

![image20210921234539686.png](https://img.learnblockchain.cn/attachments/2021/09/dBAVNhXl614ddc4a5b036.png)

其实就是在调用实际的 removeLiquidity 之前先用 permit 方式完成授权操作。

removeLiquidityETHWithPermit 也一样的，就不看代码了。

接着，来看看 removeLiquidityETHSupportingFeeOnTransferTokens 函数，先看看其代码实现：

![image20210922000049118.png](https://img.learnblockchain.cn/attachments/2021/09/aRlRZnOf614ddc7132162.png)

该函数功能和 removeLiquidityETH 一样，但对比一下，就会发现主要不同点在于：

1. 返回值没有 amountToken；
2. 调用 removeLiquidity 后也没有 amountToken 值返回
3. 进行 safeTransfer 时传值直接读取当前地址的 token 余额。

有一些项目 token，其合约实现上，在进行 transfer 的时候，就会扣减掉部分金额作为费用，或作为税费缴纳，或锁仓处理，或替代 ETH 来支付 GAS 费。总而言之，就是某些 token 在进行转账时是会产生损耗的，实际到账的数额不一定就是传入的数额。该函数主要支持的就是这类 token。

##### 2.2.3 兑换接口

兑换接口则多达 9 个：

- **swapExactTokensForTokens**：用 ERC20 兑换 ERC20，但支付的数量是指定的，而兑换回的数量则是未确定的
- **swapTokensForExactTokens**：也是用 ERC20 兑换 ERC20，与上一个函数不同，指定的是兑换回的数量
- **swapExactETHForTokens**：指定 ETH 数量兑换 ERC20
- **swapTokensForExactETH**：用 ERC20 兑换成指定数量的 ETH
- **swapExactTokensForETH**：用指定数量的 ERC20 兑换 ETH
- **swapETHForExactTokens**：用 ETH 兑换指定数量的 ERC20
- **swapExactTokensForTokensSupportingFeeOnTransferTokens**：指定数量的 ERC20 兑换 ERC20，支持转账时扣费
- **swapExactETHForTokensSupportingFeeOnTransferTokens**：指定数量的 ETH 兑换 ERC20，支持转账时扣费
- **swapExactTokensForETHSupportingFeeOnTransferTokens**：指定数量的 ERC20 兑换 ETH，支持转账时扣费

这么多个接口，我们就看看几个具有代表性的接口即可。首先是 swapExactTokensForTokens，其实现代码如下：

![image20210922132105076.png](https://img.learnblockchain.cn/attachments/2021/09/qsrwAde9614ddc9481981.png)

这是指定 amountIn 的兑换，比如用 tokenA 兑换 tokenB，那 amountIn 就是指定支付的 tokenA 的数量，而兑换回来的 tokenB 的数量自然是越多越好。

关于入参的 path，是由前端 SDK 计算出最优路径后传给合约的。至于前端是如何计算得出最优路径的，具体的算法我没去研究过前端 SDK 的实现，但在我之前写过的一篇文章[《这几天我写了一个DEX交易聚合器》](https://mp.weixin.qq.com/s?__biz=MzA5OTI1NDE0Mw==&mid=2652494325&idx=1&sn=59f0017b10da7488d4f262b6177243b7&chksm=8b6853e5bc1fdaf37b191db6eafc4e3625d09001926a2cde625775a76ad986732a2db9218069&token=1269134064&lang=zh_CN#rd)中有讲到我的一些思路，感兴趣的朋友可以去看一看。

swapExactTokensForTokens 的实现逻辑就 4 行代码而已，非常简单。第一行计算出兑换数量，第二行判断是否超过滑动计算后的最小值，第三行将支付的代币转到 pair 合约，第四行再调用兑换的内部函数。那么，再来看看这个兑换的内部函数是如何实现的：

![image20210922131920155.png](https://img.learnblockchain.cn/attachments/2021/09/5sT7VHFc614ddcc3e31a2.png)

可看到，其实现逻辑也不复杂，主要就是遍历整个兑换路径，并对路径中每两个配对的 token 调用 pair 合约的兑换函数，实现底层的兑换处理。

接着，来看看 swapTokensForExactTokens 的实现：

![image20210922153618336.png](https://img.learnblockchain.cn/attachments/2021/09/oIqsTDSJ614ddcdb5bcd1.png)

这是指定 amountOut 的兑换，比如用 tokenA 兑换 tokenB，那 amountOut 就是指定想要换回的 tokenB 的数量，而需要支付的 tokenA 的数量则是越少越好。因此，其实现代码，第一行其实就是用 amountOut 来计算得出需要多少 amountIn。返回的 amounts 数组，第一个元素就是需要支付的 tokenA 数量。其他的代码逻辑都很好理解了。

接着，来看看指定 ETH 的兑换，就以 swapExactETHForTokens 为例：

![image20210922155307659.png](https://img.learnblockchain.cn/attachments/2021/09/im42BvwG614ddcfe27000.png)

支付的 ETH 数量是从 msg.value 中读取的。而且，可看到还调用了 WETH 的 deposit 函数，将 ETH 转为了 WETH 之后再转账给到 pair 合约。这说明和前面的流动性接口一样，是将 ETH 转为 WETH 进行底层处理的。

其他几个兑换接口的逻辑也是差不多的，就不再一一讲解了。剩下的主要想聊聊支持转账时扣费的接口，就以 swapExactTokensForTokensSupportingFeeOnTransferTokens 为例，该接口的实现代码如下：

![image20210923223540938.png](https://img.learnblockchain.cn/attachments/2021/09/nNJoWT50614ddd125345b.png)

实现逻辑就只有 4 步，第一步先将 amountIn 转账给到 pair 合约，第二步读取出接收地址在兑换路径中最后一个代币的余额，第三步调用内部函数实现路径中每一步的兑换，第四步再验证接收者最终兑换得到的资产数量不能小于指定的最小值。

因为此类代币转账时可能会有损耗，所以就无法使用恒定乘积公式计算出最终兑换的资产数量，因此用交易后的余额减去交易前的余额来计算得出实际值。

而核心逻辑其实都在内部函数 _swapSupportingFeeOnTransferTokens 中实现，其代码如下：

![image20210922173948534.png](https://img.learnblockchain.cn/attachments/2021/09/8KC7FRQA614ddd242f263.png)

这里面最核心也较难理解的逻辑可能就是 amountInput 的计算，即理解好这一行代码，其他都很好理解：

```solidity
amountInput = IERC20(input).balanceOf(address(pair)).sub(reserveInput);
```

因为 input 代币转账时可能会有损耗，所以在 pair 合约里实际收到多少代币，只能通过查出 pair 合约当前的余额，再减去该代币已保存的储备量，这才能计算出来实际值。

其他代码就比较容易理解，就不展开说明了。而其他支持此类转账代币的兑换接口，和前面说的兑换接口也是类似的，所以也不一一讲解了。

##### 2.2.4 查询接口

查询接口有 5 个：

- **quote**
- **getAmountOut**
- **getAmountIn**
- **getAmountsOut**
- **getAmountsIn**

这几个查询接口的实现都是直接调用 UniswapV2Library 库对应的函数，所以也无需再赘述了。

## 三、典型项目

##### 3.1 TWAP

**TWAP = Time-Weighted Average Price**，即**时间加权平均价格**，可用来创建有效防止价格操纵的**链上价格预言机**。

TWAP 的实现机制其实很简单。首先，在配对合约里会存储三个相关变量：

- **price0CumulativeLast**
- **price1CumulativeLast**
- **blockTimestampLast**

前两个变量是两个 token 的累加价格，最后一个变量则用来记录更新的区块时间。我们可以直接来看看其代码实现：

![image20211009150104200.png](https://img.learnblockchain.cn/attachments/2021/10/YQAJGNDy616517c8e3505.png)

这是 **UniswapV2Pair** 合约的 *_update* 函数，每次 *mint*、*burn*、*swap*、*sync* 时都会触发更新。实现逻辑很容易理解，主要就以下几步：

1. 读取当前的区块时间 blockTimestamp
2. 计算出与上一次更新的区块时间之间的时间差 timeElapsed
3. 如果 timeElapsed > 0 且两个 token 的 reserve 都不为 0，则更新两个累加价格
4. 更新两个 reserve 和区块时间 blockTimestampLast

有些人可能还是不太理解累加价格的意义，要把它理解透彻，先从当前时刻的价格说起，即 token0 和 token1 的当前价格，其实可以根据以下公式计算所得：

```mathematica
price0 = reserve1 / reserve0
price1 = reserve0 / reserve1
```

比如，假设两个 token 分别为 WETH 和 USDT，当前储备量分别为 10 WETH 和 40000 USDT，那么 WETH 和 USDT 的价格分别为：

```mathematica
price0 = 40000/10 = 4000 USDT
price1 = 10/40000 = 0.00025 WETH
```

现在，再加上时间维度来考虑。比如，当前区块时间相比上一次更新的区块时间，过去了 5 秒，那就可以算出这 5 秒时间的累加价格：

```mathematica
price0Cumulative = reserve1 / reserve0 * timeElapsed = 40000/10*5 = 20000 USDT
price1Cumulative = reserve0 / reserve1 * timeElapsed = 10/40000*5 = 0.00125 WETH
```

假设之后再过了 6 秒，最新的 reserve 分别变成了 12 WETH 和 32000 USDT，则最新的累加价格变成了：

```mathematica
price0CumulativeLast = price0Cumulative + reserve1 / reserve0 * timeElapsed = 20000 + 32000/12*6 = 36000 USDT
price1CumulativeLast = price1Cumulative + reserve0 / reserve1 * timeElapsed = 0.00125 + 12/32000*6 = 0.0035 WETH
```

这就是合约里所记录的累加价格了。

另外，每次计算时因为有 timeElapsed 的判断，所以其实每次计算的是每个区块的第一笔交易。而且，计算累加价格时所用的 reserve 是更新前的储备量，所以，实际上所计算的价格是之前区块的，因此，想要操控价格的难度也就进一步加大了。

有了前面的基础，接下来就可以计算 TWAP 即时间加权平均价格了。计算公式也很简单，如下图：

![v2_twapfdc82ab82856196510db6b421cce9204.png](https://img.learnblockchain.cn/attachments/2021/10/hZCYV84U616518196ee5d.png)

代入我们的例子，为了简化，我们将前面 5 秒时间的时刻记为 T1，累加价格记为 priceT1，而 6 秒时间后的时刻记为 T2，累加价格记为 priceT2。如此，可以计算出，在后面 6 秒时间里的平均价格：

```mathematica
twap = (priceT2 - priceT1)/(T2 - T1) = (36000 - 20000)/6 = 2666.66
```

在实际应用中，一般有两种计算方案，一是固定时间窗口的 TWAP，二是移动时间窗口的 TWAP。在 uniswap-v2-periphery 项目中，examples 目录下提供了这两种方案的示例代码，分为是 **ExampleOracleSimple.sol** 和 **ExampleSlidingWindowOracle.sol**，具体代码就不展开讲解了。

现在，Uniswap TWAP 已经被广泛应用于很多 DeFi 协议，很多时候会结合 Chainlink 一起使用。比如 Compound 就使用 Chainlink 进行喂价并加入 Uniswap TWAP 进行边界校验，防止价格波动太大。

##### 3.2  FlashSwap

FlashSwap，翻译过来就是**闪电兑换**，和**闪电贷（FlashLoan）** 有点类似。

从代码层面来说，闪电兑换的触发在 **UniswapV2Pair** 合约的 **swap** 函数里的，该函数里有这么一行代码：

```solidity
if (data.length > 0) IUniswapV2Callee(to).uniswapV2Call(msg.sender, amount0Out, amount1Out, data);
```

这行代码主要说明了三个信息：

1. **to** 地址是一个合约地址
2. **to** 地址的合约实现了 **IUniswapV2Callee** 接口
3. 可以在 **uniswapV2Call** 函数里执行 **to** 合约自己的逻辑

一般情况下的兑换流程，是先支付 *tokenA*，再得到 *tokenB*。但闪电兑换却可以先得到 tokenB，最后再支付 tokenA。如下图：

![image20211010105414230.png](https://img.learnblockchain.cn/attachments/2021/10/wcus6MH9616518b5f34f5.png)

即是说，通过闪电兑换，可以实现无前置成本的套利。

比如，在 Uniswap 上可以用 3000 DAI 兑换出 1 ETH，而在 Sushi 上可以将 1 ETH 兑换成 3100 DAI，这就存在 100 DAI 的套利空间了。但是，如果用户钱包里没有 DAI 的话，该怎么套利呢？通过 Uniswap 的闪电兑换，就可以先获得 ETH，再将 ETH 在 Sushi 卖出得到 DAI，最后支付 DAI 给到 Uniswap，这样就实现了无需前置资金成本的套利了。

理论上，只要利润空间能覆盖两边的交易手续费和 GAS，就值得执行套利。这种套利行为能使得不同 DEX 之间的价格趋于一致。

闪电兑换还可以应用于另一种场景。假设用户想在 Compound 抵押 ETH 借出 DAI，再用借出的 DAI 到 Uniswap 兑换成 ETH，再抵押到 Compound 借出更多 DAI，如此重复操作，从而提高做多 ETH 的杠杆率。这么做的效率非常低。而使用闪电兑换，可以大大提高交易效率：

1. 先从 Uniswap 得到 ETH
2. 将用户的 ETH 和从 Uniswap 得到的 ETH 抵押进 Compound
3. 从 Compound 借出 DAI
4. 在 Uniswap 支付 DAI

上述步骤也不需要重复执行，一次流程就实现了用户想要的杠杆率，相比之下，明显高效很多。

在 uniswap-v2-periphery 项目中，examples 目录下有个 **ExampleFlashSwap.sol**，就是实现闪电兑换的一个示例，实现的是在 UniswapV1 和 UniswapV2 之间套利。

##### 3.3 质押挖矿

质押挖矿项目也同样很小，这是项目的 github 地址：

- [GitHub - Uniswap/liquidity-staker: Initial UNI liquidity staking contracts](https://github.com/Uniswap/liquidity-staker)

总共只有四个 sol 文件：

- **IStakingRewards.sol**
- **RewardsDistributionRecipient.sol**
- **StakingRewards.sol**
- **StakingRewardsFactory.sol**

**IStakingRewards.sol** 是一个接口文件，定义了质押合约 **StakingRewards** 需要实现的一些函数，其中，Mutative 函数只有四个：

- **stake**：充值，即质押
- **withdraw**：提现，即解质押
- **getReward**：提取奖励
- **exit**：退出

剩下的则都是 View 函数：

- **lastTimeRewardApplicable**：有奖励的最近区块数
- **rewardPerToken**：每单位 Token 奖励数量
- **earned**：用户已赚但未提取的奖励数量
- **getRewardForDuration**：挖矿奖励总量
- **totalSupply**：总质押量
- **balanceOf**：用户的质押余额

**RewardsDistributionRecipient.sol** 则是一个抽象合约，跟常用的 Ownable 合约类似，我们可以直接看看其代码实现：

![image20211010181334698.png](https://img.learnblockchain.cn/attachments/2021/10/bjjDRuNO616518e446a47.png)

总共就 12 行代码，rewardsDistribution 其实就是管理员地址，还有一个 onlyRewardsDistribution 的 modifier，这不就是和我们熟知的 Ownable 一样的功能嘛。另外，还定义了一个抽象函数 **notifyRewardAmount**，所以实际上这就是一个抽象合约。而继承了该合约的是 **StakingRewards** 合约，后面再细说。

StakingRewards.sol 留到最后再说，先来看看 **StakingRewardsFactory.sol**，这是一个工厂合约，主要就是用来部署 StakingRewards 合约的。

###### 3.3.1 StakingRewardsFactory

工厂合约里定义了四个变量：

- **rewardsToken**：用作奖励的代币，其实就是 UNI 代币
- **stakingRewardsGenesis**：质押挖矿开始的时间
- **stakingTokens**：用来质押的代币数组，一般就是各交易对的 LPToken
- **stakingRewardsInfoByStakingToken**：一个 mapping，用来保存质押代币和质押合约信息之间的映射

质押合约信息则是一个数据结构：

```solidity
struct StakingRewardsInfo {
    address stakingRewards;
    uint rewardAmount;
}
```

其中，stakingRewards 其实就是 StakingRewards 合约（即质押合约）地址，rewardAmount 则是该质押合约每周期的奖励总量。

rewardsToken 和 stakingRewardsGenesis 在工厂合约的构造函数里就初始化的。除了构造函数，工厂合约还有三个函数：

- **deploy**
- **notifyRewardAmounts**
- **notifyRewardAmount**

deploy 就是部署 StakingRewards 合约的函数，其代码实现如下：

```solidity
function deploy(address stakingToken, uint rewardAmount) public onlyOwner {
    StakingRewardsInfo storage info = stakingRewardsInfoByStakingToken[stakingToken];
    require(info.stakingRewards == address(0), 'StakingRewardsFactory::deploy: already deployed');

    info.stakingRewards = address(new StakingRewards(address(this), rewardsToken, stakingToken));
    info.rewardAmount = rewardAmount;
    stakingTokens.push(stakingToken);
}
```

两个入参，stakingToken 就是质押代币，一般为 LPToken；rewardAmount 则是奖励数量。

实现逻辑，先从 mapping 中读取出 info，如果 info 的 stakingRewards 不为零地址说明该质押代币的质押合约已经部署过了，不能重复部署。接着，用 new 的方式创建了 StakeingRewards 合约，并将合约地址赋值给 info.stakingRewards，将合约地址保存起来。之后，再保存 rewardAmount。最后，将 stakingToken 加到质押代币数组里。至此，质押合约的部署工作就完成了。

部署合约之后，下一步应该将用来挖矿的代币转入到质押合约中，这就要通过 **notifyRewardAmount** 函数了，其代码实现如下：

```solidity
function notifyRewardAmount(address stakingToken) public {
    require(block.timestamp >= stakingRewardsGenesis, 'StakingRewardsFactory::notifyRewardAmount: not ready');

    StakingRewardsInfo storage info = stakingRewardsInfoByStakingToken[stakingToken];
    require(info.stakingRewards != address(0), 'StakingRewardsFactory::notifyRewardAmount: not deployed');

    if (info.rewardAmount > 0) {
        uint rewardAmount = info.rewardAmount;
        info.rewardAmount = 0;

        require(
        IERC20(rewardsToken).transfer(info.stakingRewards, rewardAmount),
        'StakingRewardsFactory::notifyRewardAmount: transfer failed'
        );
        StakingRewards(info.stakingRewards).notifyRewardAmount(rewardAmount);
    }
}
```

调用该函数之前，其实还有一个前提条件要先完成，那就是**需要先将用来挖矿奖励的 UNI 代币数量先转入该工厂合约**。有个这个前提，工厂合约的该函数才能实现将 UNI 代币下发到质押合约中去。

代码逻辑就很简单了，先是判断当前区块的时间需大于等于质押挖矿的开始时间。然后读取出指定的质押代币 stakingToken 映射的质押合约 info，要求 info 的质押合约地址不能为零地址，否则说明还没部署。再判断 info.rewardAmount 是否大于零，如果为零也不用下发奖励。if 语句里面的逻辑主要就是调用 rewardsToken 的 transfer 函数将奖励代币转发给质押合约，再调用质押合约的 notifyRewardAmount 函数触发其内部处理逻辑。另外，将 info.rewardAmount 重置为 0，可以避免向质押合约重复下发奖励代币。

而 **notifyRewardAmounts** 函数，则是遍历整个质押代币数组，对每个代币再调用 **notifyRewardAmount**，实现逻辑非常简单。

至此，工厂合约的代码逻辑就讲完了。下面，就来看看 StakingRewards 合约了。

###### 3.3.2 StakingRewards

**StakingRewards** 合约会继承 **RewardsDistributionRecipient** 合约和 **IStakingRewards** 接口。

StakingRewards 存储的变量则比较多，除了继承自 **RewardsDistributionRecipient** 抽象合约里的 rewardsDistribution 变量之外，还有 11 个变量：

- **rewardsToken**：奖励代币，即 UNI 代币
- **stakingToken**：质押代币，即 LPToken
- **periodFinish**：质押挖矿结束的时间，默认时为 0
- **rewardRate**：挖矿速率，即每秒挖矿奖励的数量
- **rewardsDuration**：挖矿时长，默认设置为 60 天
- **lastUpdateTime**：最近一次更新时间
- **rewardPerTokenStored**：每单位 token 奖励数量
- **userRewardPerTokenPaid**：用户的每单位 token 奖励数量
- **rewards**：用户的奖励数量
- **_totalSupply**：私有变量，总质押量
- **_balances**：私有变量，用户质押余额

前面讲工厂合约的 notifyRewardAmount 函数时，提到最后其实会调用到 StakingRewards 合约的 notifyRewardAmount 函数，我们就来看看这个函数是如何实现的：

```solidity
function notifyRewardAmount(uint256 reward) external onlyRewardsDistribution updateReward(address(0)) {
    if (block.timestamp >= periodFinish) {
        rewardRate = reward.div(rewardsDuration);
    } else {
        uint256 remaining = periodFinish.sub(block.timestamp);
        uint256 leftover = remaining.mul(rewardRate);
        rewardRate = reward.add(leftover).div(rewardsDuration);
    }

    // Ensure the provided reward amount is not more than the balance in the contract.
    // This keeps the reward rate in the right range, preventing overflows due to
    // very high values of rewardRate in the earned and rewardsPerToken functions;
    // Reward + leftover must be less than 2^256 / 10^18 to avoid overflow.
    uint balance = rewardsToken.balanceOf(address(this));
    require(rewardRate <= balance.div(rewardsDuration), "Provided reward too high");

    lastUpdateTime = block.timestamp;
    periodFinish = block.timestamp.add(rewardsDuration);
    emit RewardAdded(reward);
}
```

该函数由工厂合约触发执行，而且根据工厂合约的代码逻辑，该函数也只会被触发一次。

由于 **periodFinish** 默认值为 0 且只会在该函数中更新值，所以只会执行 **block.timestamp >= periodFinish** 的分支逻辑，将从工厂合约转过来的挖矿奖励总量除以挖矿奖励时长，得到挖矿速率 **rewardRate**，即每秒的挖矿数量。理论上，else 分支是执行不到的，除非以后工厂合约升级为可以多次触发执行该函数。之后，读取 balance 并校验下 rewardRate，可以保证收取到的挖矿奖励余额也是充足的，rewardRate 就不会虚高。最后，更新 **lastUpdateTime** 和 **periodFinish**。periodFinish 就是在当前区块时间上加上挖矿时长，就得到了挖矿结束的时间。

接着，再来看看几个核心业务函数的实现，包括 stake、withdraw、getReward。

**stake** 就是质押代币的函数，实现代码如下：

```solidity
function stake(uint256 amount) external nonReentrant updateReward(msg.sender) {
    require(amount > 0, "Cannot stake 0");
    _totalSupply = _totalSupply.add(amount);
    _balances[msg.sender] = _balances[msg.sender].add(amount);
    stakingToken.safeTransferFrom(msg.sender, address(this), amount);
    emit Staked(msg.sender, amount);
}
```

函数体内的代码逻辑很简单，将用户指定的质押量 amount 增加到 _totalSupply（总质押量）和 _balances（用户的质押余额），最后调用 stakingToken 的 safeTransferFrom 将代币从用户地址转入当前合约地址。

**withdraw** 则是用来提取质押代币的，代码实现也同样很简单，_totalSupply 和 _balances 都减掉提取数量，且将代币从当前合约地址转到用户地址：

```solidity
function withdraw(uint256 amount) public nonReentrant updateReward(msg.sender) {    
    require(amount > 0, "Cannot withdraw 0");
    _totalSupply = _totalSupply.sub(amount);
    _balances[msg.sender] = _balances[msg.sender].sub(amount);
    stakingToken.safeTransfer(msg.sender, amount);
    emit Withdrawn(msg.sender, amount);
}
```

**getReward** 是领取挖矿奖励的函数，内部逻辑主要就是从 rewards 中读取出用户有多少奖励并清零和转账给到用户：

```solidity
function getReward() public nonReentrant updateReward(msg.sender) {
    uint256 reward = rewards[msg.sender];
    if (reward > 0) {
        rewards[msg.sender] = 0;
        rewardsToken.safeTransfer(msg.sender, reward);
        emit RewardPaid(msg.sender, reward);
    }
}
```

这几个核心业务函数体内的逻辑都非常好理解，值得一说的其实是每个函数声明最后的 **updateReward(msg.sender)**，这是一个更新挖矿奖励的 modifer，我们来看其代码：

```solidity
modifier updateReward(address account) {
    rewardPerTokenStored = rewardPerToken();
    lastUpdateTime = lastTimeRewardApplicable();
    if (account != address(0)) {
        rewards[account] = earned(account);
        userRewardPerTokenPaid[account] = rewardPerTokenStored;
    }
    _;
}
```

主要逻辑就是更新几个字段，包括 rewardPerTokenStored、lastUpdateTime 和用户的奖励相关的 rewards[account] 和 userRewardPerTokenPaid[account]。

其中，还调用到其他三个函数：rewardPerToken()、lastTimeRewardApplicable()、earned(account)。先来看看这三个函数的实现。最简单的就是 lastTimeRewardApplicable：

```solidity
function lastTimeRewardApplicable() public view returns (uint256) {
    return Math.min(block.timestamp, periodFinish);
}
```

其逻辑就是从**当前区块时间**和**挖矿结束时间**两者中返回最小值。因此，当挖矿未结束时返回的就是当前区块时间，而挖矿结束后则返回挖矿结束时间。也因此，挖矿结束后，lastUpdateTime 也会一直等于挖矿结束时间，这点很关键。

rewardPerToken 函数则是获取每单位质押代币的奖励数量，其实现代码如下：

```solidity
function rewardPerToken() public view returns (uint256) {
    if (_totalSupply == 0) {
        return rewardPerTokenStored;
    }
    return
        rewardPerTokenStored.add(
            lastTimeRewardApplicable().sub(lastUpdateTime).mul(rewardRate).mul(1e18).div(_totalSupply)
        );
}
```

这其实就是用累加计算的方式存储到 rewardPerTokenStored 变量中。当挖矿结束后，则不会再产生增量，rewardPerTokenStored 就不会再增加了。

earned 函数则是计算用户当前的挖矿奖励，代码实现也只有一行代码：

```solidity
function earned(address account) public view returns (uint256) {
    return _balances[account].mul(rewardPerToken().sub(userRewardPerTokenPaid[account])).div(1e18).add(rewards[account]);
}
```

其逻辑也是计算出增量的每单位质押代币的挖矿奖励，再乘以用户的质押余额得到增量的总挖矿奖励，再加上之前已存储的挖矿奖励，就得到当前总的挖矿奖励。

至此，StakingRewards 合约的主要实现逻辑也都讲解完了

## 总结

至此，所有 UniswapV2 的合约项目就都讲解完了。虽然分为了好几个小项目，但从架构设计上来说，能够大大减低不同模块之间的耦合性，不同项目也可以由不同的小团队单独维护，而且项目小而简单，那出 BUG 的概率也会更低。所以，这样的架构设计其实更适合 Dapp。
