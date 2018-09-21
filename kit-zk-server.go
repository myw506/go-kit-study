package main

import (
	"context"
	"grpc-sample/pb"
	"net"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	//	"github.com/go-kit/kit/sd/etcdv3"
	stdzk "github.com/go-kit/kit/sd/zk"
	grpc_transport "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
)

type BookServer struct {
	bookListHandler grpc_transport.Handler
	bookInfoHandler grpc_transport.Handler
}

//通过grpc调用GetBookInfo时,GetBookInfo只做数据透传, 调用BookServer中对应Handler.ServeGRPC转交给go-kit处理
func (s *BookServer) GetBookInfo(ctx context.Context, in *book.BookInfoParams) (*book.BookInfo, error) {
	_, rsp, err := s.bookInfoHandler.ServeGRPC(ctx, in)
	if err != nil {
		return nil, err
	}
	return rsp.(*book.BookInfo), err
}

//通过grpc调用GetBookList时,GetBookList只做数据透传, 调用BookServer中对应Handler.ServeGRPC转交给go-kit处理
func (s *BookServer) GetBookList(ctx context.Context, in *book.BookListParams) (*book.BookList, error) {
	_, rsp, err := s.bookListHandler.ServeGRPC(ctx, in)
	if err != nil {
		return nil, err
	}
	return rsp.(*book.BookList), err
}

//创建bookList的EndPoint
func makeGetBookListEndpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		//请求列表时返回 书籍列表
		bl := new(book.BookList)
		bl.BookList = append(bl.BookList, &book.BookInfo{BookId: 1, BookName: "21天精通php"})
		bl.BookList = append(bl.BookList, &book.BookInfo{BookId: 2, BookName: "21天精通java"})
		return bl, nil
	}
}

//创建bookInfo的EndPoint
func makeGetBookInfoEndpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		//请求详情时返回 书籍信息
		req := request.(*book.BookInfoParams)
		b := new(book.BookInfo)
		b.BookId = req.BookId
		b.BookName = "21天精通php"
		return b, nil
	}
}

func decodeRequest(_ context.Context, req interface{}) (interface{}, error) {
	return req, nil
}

func encodeResponse(_ context.Context, rsp interface{}) (interface{}, error) {
	return rsp, nil
}

func main() {
	var (
		//etcd服务地址
		zkServer = "192.168.0.140:2181"

		//服务的信息目录
		Path = "/services/"
		//当前启动服务实例的地址
		instance    = "127.0.0.1:50052"
		serviceName = "book"
		//服务实例注册的路径
		//		key = prefix + instance

		//服务实例注册的val
		//		value = instance
		//		ctx = context.Background()
		//服务监听地址
		serviceAddress = ":50052"
	)

	//etcd的连接参数
	//	options := etcdv3.ClientOptions{
	//		DialTimeout:   time.Second * 3,
	//		DialKeepAlive: time.Second * 3,
	//	}

	//	options := stdzk.clientConfig{
	//		connectTimeout: time.Second * 3,
	//		sessionTimeout: time.Second * 5,
	//	}

	//创建etcd连接
	//	client, err := etcdv3.NewClient(ctx, []string{etcdServer}, options)
	client, err := stdzk.NewClient([]string{zkServer}, log.NewNopLogger())

	if err != nil {
		panic(err)
	}

	defer client.Stop()

	// 创建注册器
	//	registrar := etcdv3.NewRegistrar(client, etcdv3.Service{
	//		Key:   key,
	//		Value: value,
	//	}, log.NewNopLogger())

	//	Path string // discovery namespace, example: /myorganization/myplatform/
	//	Name string // service name, example: addscv
	//	Data []byte // instance data to store for discovery, example: 10.0.2.10:80

	registrar := stdzk.NewRegistrar(client, stdzk.Service{
		Path: Path,
		Name: serviceName,
		Data: []byte(instance),
	}, log.NewNopLogger())

	// 注册器启动注册
	registrar.Register()

	bookServer := new(BookServer)

	bookListHandler := grpc_transport.NewServer(
		makeGetBookListEndpoint(),
		decodeRequest,
		encodeResponse,
	)
	bookServer.bookListHandler = bookListHandler

	bookInfoHandler := grpc_transport.NewServer(
		makeGetBookInfoEndpoint(),
		decodeRequest,
		encodeResponse,
	)
	bookServer.bookInfoHandler = bookInfoHandler

	ls, _ := net.Listen("tcp", serviceAddress)

	gs := grpc.NewServer(grpc.UnaryInterceptor(grpc_transport.Interceptor))
	book.RegisterBookServiceServer(gs, bookServer)
	gs.Serve(ls)
}
