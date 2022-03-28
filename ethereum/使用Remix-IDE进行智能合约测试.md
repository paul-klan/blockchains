### 使用Remix-IDE进行智能合约测试

文章来源：https://learnblockchain.cn/article/809

- [测试](https://learnblockchain.cn/tags/%E6%B5%8B%E8%AF%95)
   - [智能合约](https://learnblockchain.cn/tags/%E6%99%BA%E8%83%BD%E5%90%88%E7%BA%A6)
   - [Solidity](https://learnblockchain.cn/tags/Solidity)
   - [Remix](https://learnblockchain.cn/tags/Remix)

IDE开发工具的简单使用，通过完成一个合约测试实例，快速上手Remix。

## 一、简介

Remix-IDE 是一个在线智能合约开发的网站，包含一些运行环境，可以在线上直接编写合约脚本，并且进行合约测试。  
本文完成通过完成一个智能合约脚本的测试过程，来达到熟练掌握Remix-IDE 以及智能合约的开发以及测试的目的。

## 二、Remix-IDE布局

![image](https://img.learnblockchain.cn/2020/03/27_/964150452.png)

先简单看下界面中的每个部分有什么作用：

1. 图标面板（IconPanel）

单击以更改哪个插件显示在侧面板中，这里面从上到下依次有：  
文件浏览、切换脚本语言用Vyper、切换脚本语言用Solidity、运行换部署合约脚本、脚本的金泰分析（检测语法错误）、单元测试、插件管理

2. 侧面板（SidePanel）

大多数（但不是全部）插件将在此处展示它的操作界面。如果你点击"文件浏览"，这里会显示当前脚本文件。

3. 主面板（MainPanel）

主要用于编辑文件。在选项卡中可以是用于IDE编译的插件或文件。

4. 终端（显示执行结果）

您将在其中查看与GUI交互的结果。您也可以在此处运行脚本。

## 三、运行环境

![image](https://img.learnblockchain.cn/2020/03/27_/961573252.png)

RemixIDE 包含两种智能合约脚本语言环境，Solidity 和 Vyper。

如果你喜欢暗黑色的背景，你可以在这里设置：  
![image](https://img.learnblockchain.cn/2020/03/27_/72864717.png)

## 四、智能合约编写

与Python类似，这两种脚本语言的  
执行文件过程也基本相同。

虽然Python脚本执行为

```
python file_name.py
```

使用vyper编译脚本

```
vyper file_name.vy
```

使用solidity编译脚本

```
solc file_name.sol
```

---

此处，我们使用Solidity编写一个简单的智能合约，现在就算你都不理解也不要紧，后面我们会有逐行的讲解：

**Solidity智能合约 （一）**

```
pragma solidity  >=0.4.22 <0.7.0;

contract SimpleStorage {
    uint storedData;

    function set(uint x) public {
        storedData = x;
    }

    function get() public view returns (uint) {
        return storedData;
    }
}
```

逐行讲解：

```
pragma solidity >=0.4.22 <0.7.0;
//第一行就是告诉大家源代码
//使用Solidity版本大于0.4.0
//并且小于 0.7.0
contract SimpleStorage {}


// contract 说明这是一个合约
// SimpleStorage 是合约的名字，叫做“简单存储”
// 这个合约实现一个功能，就是
// 用户将一个数字存储到区块链
// 数据库中，其他用户可以访问

 uint storedData;

// 定义一个变量 storedData

 function set(uint x) public {
        storedData = x;
    }

// set() 函数，输入一个值 x
// 把x赋值给 storedData
// 完成数据的存储

  function get() public view returns (uint) {
        return storedData;
    }

// get() 函数，没有输入值
// 直接获取变量storedData的值
// 返回storedData的值
```

该合约能完成的事情并不多（由于以太坊构建的基础架构的原因）：它能允许任何人在合约中存储一个单独的数字，并且这个数字可以被世界上任何人访问，且没有可行的办法阻止你发布这个数字。当然，任何人都可以再次调用 set ，传入不同的值，覆盖你的数字，但是这个数字仍会被存储在区块链的历史记录中。随后，我们会看到怎样施加访问限制，以确保只有你才能改变这个数字。

简单的讲： **这个智能合约帮你存一个数到区块链中。**

关于测试：

如何测试这个合约是否正确，那么就看用户set（x）写入到区块链中的数字，是否与他get()到的数字x相同。就知道是否数据写入的正确无误。

## 五、执行智能合约

在RemixIDE 文件浏览中，点击+图标，添加一个文件`Demo.sol`  
将上面的合约脚本代码复制到该文件中。

一般会自动编译，编译报错会爆红色信息，否则在左侧，Solidity图标处，可以看到编译成功的√对勾。  
![image](https://img.learnblockchain.cn/2020/03/27_/374877172.png)  
点击侧边栏中部署图标，进行账户地址的相关配置，就可以将智能合约部署在区块链中。

但是在这之前，也就是本文的关键，我们需要对这个脚本进行测试。即：编写智能合约测试脚本，并执行测试。

## 六、编写智能合约测试脚本

点击左侧，![image](https://img.learnblockchain.cn/2020/03/27_/40515220.png)进入单元测试。

内容如下：  
![image](https://img.learnblockchain.cn/2020/03/27_/684942712.png)

点击`Generate test file`生成测试脚本。  
生成的后缀是'_test'的测试文件的基本模板。  
改名为`Demo_test.sol`,并编写如下内容进行测试：

```
pragma solidity >=0.4.22 <0.7.0;
import "remix_tests.sol"; // this import is automatically injected by Remix.
import "./Demo.sol";

contract DemoTest {

    SimpleStorage simpleStorage;  //定义合约变量     

    function beforeAll () public {

         simpleStorage = new SimpleStorage();   // 创建一个合约对象  simpleStorage

    }

    function checkNumberRight () public {
        // 使用 set（）函数设定 x数值为10 ，将其写入了simpleStorage中 
        simpleStorage.set(uint(10));

        Assert.equal(simpleStorage.get(), uint(10), "simpleStorage == 10");
    }


}
```

先勾选`RunTests`下面的测试文件，然后点击`RunTests`进行合约测试。

需要了解的基本的合约测试知识：

```
除此之外，Remix还允许使用一些特殊功能来使测试更具结构性。他们是：

beforeEach() -每次测试前运行
beforeAll() -在所有测试之前运行
afterEach() -每次测试后运行
afterAll() -在所有测试后运行
```

这个脚本，主要用到了`beforeAll ()`在测试之前来创建自己合约的对象。

## 七、智能合约测试

逐行解释测试脚本：

```
pragma solidity >=0.4.22 <0.7.0;  // 如上所述， 进行solidity 版本的设定
import "remix_tests.sol"; //引入 Remix自动测试的库 默认写即可
import "./Demo.sol";  // 重要： 引入你要测试的“智能合约”

contract DemoTest {   // 测试函数命名

    SimpleStorage simpleStorage;  //定义合约变量     


    // 该函数在执行测试之前执行，先把你的合约创建并赋值给一个对象
    function beforeAll () public {
         simpleStorage = new SimpleStorage();   
         // 创建一个合约对象  simpleStorage
    }


    // 重要：测试函数
    function checkNumberRight () public {
        // 使用 set（）函数设定 x数值为10 ，将其写入了simpleStorage中  
        simpleStorage.set(uint(10));

        // 现在测试simpleStorage的值是否等于 你输入的值10 
        // 用 get()函数获取simpleStorage的值
        // uint(10)  表示数字 10
        // Assert.equal() 表示是否两者相等，不相等，报错提示，测试不通过
        Assert.equal(simpleStorage.get(), uint(10), "simpleStorage != 10");
    }


}
```

了解并完成了测试脚本后，并勾选了对应的文件，就可以点击`RunTests`运行脚本测试：  
![image](https://img.learnblockchain.cn/2020/03/27_/268180639.png)

结果如上，则表示通过测试脚本。

## 八、网站地址

[RemixIDE 官网](https://remix.ethereum.org/)

[RemixIDE-中文](http://remix.hubwiz.com/)

[Solidity-文档](https://learnblockchain.cn/docs/solidity/)

[Vyper-文档](https://vyper.readthedocs.io/en/latest/)

本文参与[登链社区写作激励计划](https://learnblockchain.cn/site/coins) ，好文好收益，欢迎正在阅读的你也加入。

-  发表于 2020-03-27 10:45
   - 阅读 ( 3569 )
   - 学分 ( 99 )
   - 分类：[智能合约](https://learnblockchain.cn/categories/%E6%99%BA%E8%83%BD%E5%90%88%E7%BA%A6)
