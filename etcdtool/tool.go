package etcdtool

import (
	"context"
	"runtime"
	"time"

	"github.com/pkg/errors"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	EtcdKeyDelete = "DELETE"
	EtcdKeyCreate = "CREATE"
	EtcdKeyModify = "MODIFY"
)

type EtcdTool struct {
	Tool *clientv3.Client
}

// leaseId 租约id ，保活使用
type WatchHandleCallBack func(key string, leaseId int64, tool *EtcdTool)

func GetEtcdTool() *EtcdTool {
	return tool
}

// 监控一个key 发送变化 回调，
// status 取值 EtcdKeyDelete ,EtcdKeyCreate , EtcdKeyModify
type DidChangeFunc func(status, key, value string)

// put 值
func (etcd *EtcdTool) PutWithTimeOut(key, value string, timeout int) (err error) {
	if key == "" {
		err = errors.Errorf("PutWithTimeOut_err key is empty ")
		return
	}
	lease, e := etcdcli.Grant(context.Background(), int64(timeout))
	if e != nil {
		err = errors.Errorf("PutWithTimeOut_err grant err %+v\n ", e)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	_, e = etcdcli.Put(ctx, key, value, clientv3.WithLease(lease.ID))
	if e != nil {
		err = errors.Errorf("PutWithTimeOut_err %+v \n", e)
		return
	}
	return
}

// put
func (etcd *EtcdTool) PutWithLeaseId(key, value string, leaseId int64) (err error) {
	if key == "" {
		err = errors.Errorf("PutWithLeaseId_err key is empty ")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	_, e := etcdcli.Put(ctx, key, value, clientv3.WithLease(clientv3.LeaseID(leaseId)))
	if e != nil {
		err = errors.Errorf("PutValue_err %+v \n", e)
		return
	}
	return
}

// put
func (etcd *EtcdTool) Put(key, value string) (err error) {
	if key == "" {
		err = errors.Errorf("PutValue_err key is empty ")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	_, e := etcdcli.Put(ctx, key, value)
	if e != nil {
		err = errors.Errorf("PutValue_err %+v \n", e)
		return
	}
	return
}

func (etcd *EtcdTool) Delete(key string) (err error) {
	if key == "" {
		err = errors.Errorf("Delete_err key is empty ")
		return
	}
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*10)
	defer cancel()
	_, e := etcdcli.Delete(ctx, key)
	if e != nil {
		err = errors.Wrapf(e, "Delete_err ")
		return
	}
	return
}

func (etcd *EtcdTool) DeletePrefix(prefixKey string) (datas []string, err error) {
	if prefixKey == "" {
		err = errors.Errorf("DeletePrefix_err key is empty ")
		return
	}
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*10)
	defer cancel()

	_, e := etcdcli.Delete(ctx, prefixKey, clientv3.WithPrefix())
	if e != nil {
		err = errors.Wrapf(e, "GetPrefixValue_err")
		return
	}

	return
}

// 失败返回err ,成功返回数据 或者空字符
func (etcd *EtcdTool) Get(key string) (data string, err error) {
	if key == "" {
		err = errors.Errorf("GetValue_err key is empty ")
		return
	}
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*10)
	defer cancel()

	resp, e := etcdcli.Get(ctx, key)
	if e != nil {
		err = errors.Wrapf(e, "GetValue_err")
		return
	}
	// 未获取到数据
	if resp == nil || resp.Kvs == nil {
		return
	}
	data = string(resp.Kvs[0].Value)
	return
}

// 失败返回err ,成功返回数据 或者 返回nil
func (etcd *EtcdTool) GetPrefix(prefixKey string) (datas []string, err error) {
	if prefixKey == "" {
		err = errors.Errorf("GetPrefixValue_err key is empty ")
		return
	}
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*10)
	defer cancel()

	resp, e := etcdcli.Get(ctx, prefixKey, clientv3.WithPrefix())
	if e != nil {
		err = errors.Wrapf(e, "GetPrefixValue_err")
		return
	}
	// 未获取到数据
	if resp == nil || resp.Kvs == nil || len(resp.Kvs) == 0 {
		return
	}
	tmp := make([]string, 0, 100)
	for i := 0; i < len(resp.Kvs); i++ {
		value := string(resp.Kvs[i].Value)
		tmp = append(tmp, value)
	}

	if len(tmp) > 0 {
		datas = tmp
	}

	return
}

// 监控一个带前缀的key
func (etcd *EtcdTool) WatchPrefix(prefixKey string, ctx context.Context, changeFunc DidChangeFunc) (err error) {
	if prefixKey == "" || ctx == nil {
		err = errors.Errorf("WatchPrefix_err args err prefixkey = %s , ctx = %+v \n", prefixKey, ctx)
		return
	}
	go func() {
		watchChan := etcdcli.Watch(context.TODO(), prefixKey, clientv3.WithPrefix())
		for {
			select {
			case data := <-watchChan:
				handleEvents(data, "WatchPrefix_trace", changeFunc)
			case <-ctx.Done():
				//logger.Info("WatchPrefix_info watch key finished prefixKey", prefixKey)
				return
			}
			runtime.Gosched()
		}
	}()

	return
}

func handleEvents(data clientv3.WatchResponse, logPrefix string, changeFunc DidChangeFunc) {
	for i := 0; i < len(data.Events); i++ {
		event := data.Events[i]
		key := string(event.Kv.Key)
		value := string(event.Kv.Value)
		status := ""
		// 删除事件
		if event.Type == mvccpb.DELETE && changeFunc != nil {
			status = EtcdKeyDelete
			changeFunc(status, key, value)
			//str := fmt.Sprintf("etcd监控删除事件 事件类型 = %s ,  key = %s , value = %s \n", status, string(event.Kv.Key), string(event.Kv.Value))
		}
		// 创建事件
		if event.Type == mvccpb.PUT && event.IsCreate() == true && changeFunc != nil {
			status = EtcdKeyCreate
			changeFunc(status, key, value)
			//str := fmt.Sprintf("etcd监控创建事件 事件类型 = %s ,  key = %s , value = %s \n", status, string(event.Kv.Key), string(event.Kv.Value))
			//logger.Debug(pre, logPrefix, "", str)
		}
		// 更新事件
		if event.Type == mvccpb.PUT && event.IsModify() == true && changeFunc != nil {
			status = EtcdKeyModify
			changeFunc(status, key, value)
		}
	}
}

// 监控一个带前缀的key
func (etcd *EtcdTool) WatchKey(watchKey string, ctx context.Context, changeFunc DidChangeFunc) (err error) {
	//pre := "WatchKey"
	if watchKey == "" || ctx == nil {
		err = errors.Errorf("WatchKey_err args err prefixkey = %s , ctx = %+v \n", watchKey, ctx)
		return
	}
	go func() {
		watchChan := etcdcli.Watch(context.TODO(), watchKey)
		for {
			select {
			case data := <-watchChan:
				handleEvents(data, "WatchKey_trace", changeFunc)
			case <-ctx.Done():
				//str := fmt.Sprintf("WatchPrefix_info watch key finished prefixKey = %s \n", watchKey)
				//logger.Info(pre, str)
				return
			}
			runtime.Gosched()
		}

	}()

	return
}

// 注册一个key 并且 使用租约保活
func (etcd *EtcdTool) RegistKeyAndKeepAlive(key, value string, ctx context.Context, callBack WatchHandleCallBack) (err error) {
	if key == "" || ctx == nil {
		err = errors.New("RegistKeyAndKeepAlive_err key or context is empty ")
		return
	}
	//pre := "RegistKeyAndKeepAlive"
	lease, e := etcd.Tool.Grant(context.TODO(), 6)
	if e != nil || lease == nil {
		err = errors.Wrapf(e, "RegistKeyAndKeepAlive_err")
		//logger.Error(err, pre, fmt.Sprintf("RegistKeyAndKeepAlive_err create lease err %+v , lease %+v \n", e, lease))
		return
	}
	//str := fmt.Sprintf("RegistKeyAndKeepAlive_info create lease succ lease.id %d , key %s , value = %s  \n", lease.ID, key, value)
	//logger.Info(pre, str)
	_, e = etcd.Tool.Put(context.TODO(), key, value, clientv3.WithLease(lease.ID))
	if e != nil {
		err = errors.Wrapf(e, "RegistKeyAndKeepAlive_err put value err %+v , key %s \n", e, key)
		//str := fmt.Sprintf("RegistKeyAndKeepAlive_err put value err %+v , key %s \n", e, key)
		//logger.Error(err, str)
		return
	}

	watchChan, e := etcd.Tool.KeepAlive(context.TODO(), lease.ID)
	if e != nil {
		err = errors.Wrapf(e, "RegistKeyAndKeepAlive_err")
		//str := fmt.Sprintf("RegistKeyAndKeepAlive_err keepalive err %+v , key %s \n", err, key)
		//logger.Error(err, str)
		return
	}

	go func() {
		for {
			select {
			case resp := <-watchChan:
				if resp != nil && callBack != nil {
					callBack(key, int64(resp.ID), etcd)
				}
			case <-ctx.Done():
				//logger.Info(pre, fmt.Sprintf("RegistKeyAndKeepAlive_info watch key end key is %s, lease.id %d\n", key, lease.ID))
				return
			}
		}
	}()

	return
}
