(1)
op etcd比较常见的写法
适用于可选条件，且参数比较多的时候可以用
但不适用于同时用多个方法(可读性差)

(2)聚合的interface写法

(3)总分总形式，方法聚合有可以分离
//Put、Get、Delete都是用Do,结构返回统一用OpResponse,又可以获取具体的，如Put put := op.Put()
