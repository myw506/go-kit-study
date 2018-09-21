package main

import (
	"context"
	"fmt"
	"grpc-sample/pb"

	"google.golang.org/grpc"
)

func main() {
	serviceAddress := "127.0.0.1:50052"
	conn, err := grpc.Dial(serviceAddress, grpc.WithInsecure())
	if err != nil {
		panic("connect error")
	}
	defer conn.Close()

	bookClient := book.NewBookServiceClient(conn)
	bi, _ := bookClient.GetBookInfo(context.Background(), &book.BookInfoParams{BookId: 1})

	fmt.Println("获取书籍详情")
	fmt.Println("bookId: 1", " => ", "bookName:", bi.BookName)

	bl, _ := bookClient.GetBookList(context.Background(), &book.BookListParams{Page: 1, Limit: 10})
	fmt.Println("获取书籍列表")

	for _, b := range bl.BookList {
		fmt.Println("bookId:", b.BookId, " => ", "bookName:", b.BookName)
	}
}
