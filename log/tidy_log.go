/**
* @Author: TongTongLiu
* @Date: 2021/3/11 4:49 PM
**/


package log

// 将 plan 、flow、 case 的状态设置为终止
// 比如 siber 异常重启，此时 case 的状态是 "进行中"，应该订正为 "终止"



// 日志归档：比如将 180 天以前的 case，迁移至归档集合
// 避免数据量过大，引起 OLTP 性能问题