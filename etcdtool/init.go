package etcdtool

var etcdcli *clientv3.Client
var tool *EtcdTool

func InitEtcd(cli *clientv3.Client) {
	if cli == nil {
		panic("etcd init fail")
	}
	etcdcli = cli
	tool = &EtcdTool{Tool: etcdcli}
}
