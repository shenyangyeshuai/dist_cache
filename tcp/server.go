package tcp

import (
	"../cache"
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
)

type Server struct {
	c cache.Cache
}

func New(c cache.Cache) *Server {
	return &Server{c: c}
}

func (s *Server) Listen() {
	l, err := net.Listen("tcp", ":12346")
	if err != nil {
		panic(err)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}
		go s.process(conn)
	}
}

type result struct {
	v []byte
	e error
}

func reply(conn net.Conn, resultChan chan chan *result) {
	defer conn.Close()

	for {
		// 收到一个 channel
		c, open := <-resultChan
		if !open {
			return
		}

		r := <-c // 等待缓存操作的结果
		e := sendResponse(r.v, r.e, conn)
		if e != nil {
			log.Println("close connection due to error:", e)
			return
		}
	}
}

func (s *Server) process(conn net.Conn) {
	r := bufio.NewReader(conn)

	resultChan := make(chan chan *result, 5000)
	defer close(resultChan)

	go reply(conn, resultChan)

	for { // for 循环的目的就是当一切都没出错的话, 不用退出(当然也不用关闭 conn), 继续等待下一次的操作
		op, err := r.ReadByte() // 阻塞式等待
		if err != nil {
			// 不是读完了, 而是出错了, 记录日志
			if err != io.EOF {
				log.Println("close connection due to error:", err)
				return
			}
			// 读完了(实际上应该是什么都没读到, 原因时客户端退出了), 就直接返回(这是退出条件)
			return
		}

		// 后续处理也有可能出错
		if op == 'S' {
			s.set(resultChan, r)
		} else if op == 'G' {
			s.get(resultChan, r)
		} else if op == 'D' {
			s.del(resultChan, r)
		} else {
			// 后续处理本身也有可能出错
			log.Println("close connection due to invalid operation:", op)
			return
		}
	}
}

func (s *Server) set(resultChan chan chan *result, r *bufio.Reader) {
	c := make(chan *result)
	resultChan <- c

	k, v, e := s.readKeyAndValue(r)
	if e != nil {
		c <- &result{v: nil, e: e}
		return
	}

	go func() {
		c <- &result{v: nil, e: s.c.Set(k, v)}
	}()
}

func (s *Server) get(resultChan chan chan *result, r *bufio.Reader) {
	c := make(chan *result)
	resultChan <- c

	k, e := s.readKey(r)
	if e != nil {
		c <- &result{v: nil, e: e}
		return
	}

	go func() {
		v, e := s.c.Get(k)
		c <- &result{v, e}
	}()
}

func (s *Server) del(resultChan chan chan *result, r *bufio.Reader) {
	c := make(chan *result)
	resultChan <- c

	k, e := s.readKey(r)
	if e != nil {
		c <- &result{v: nil, e: e}
		return
	}

	go func() {
		c <- &result{v: nil, e: s.c.Del(k)}
	}()
}

func sendResponse(value []byte, err error, conn net.Conn) error {
	if err != nil {
		errStr := err.Error()
		// 错误的话给客户端写回 "- 错误信息"
		tmp := fmt.Sprintf("-%d %s", len(errStr), errStr)
		_, e := conn.Write([]byte(tmp))
		return e
	}

	vlen := fmt.Sprintf("%d ", len(value))
	// 正确的话给客户端写回 "数据长度 数据内容"
	_, e := conn.Write(append([]byte(vlen), value...))
	return e
}

func (s *Server) readKey(r *bufio.Reader) (string, error) {
	klen, e := s.readLen(r)
	if e != nil {
		return "", e
	}

	k := make([]byte, klen)
	_, e = io.ReadFull(r, k)
	if e != nil {
		return "", e
	}

	return string(k), nil
}

func (s *Server) readKeyAndValue(r *bufio.Reader) (string, []byte, error) {
	klen, e := s.readLen(r)
	if e != nil {
		return "", nil, e
	}

	vlen, e := s.readLen(r)
	if e != nil {
		return "", nil, e
	}

	k := make([]byte, klen)
	_, e = io.ReadFull(r, k)
	if e != nil {
		return "", nil, e
	}

	v := make([]byte, vlen)
	_, e = io.ReadFull(r, v)
	if e != nil {
		return "", nil, e
	}

	return string(k), v, nil
}

func (s *Server) readLen(r *bufio.Reader) (int, error) {
	tmp, e := r.ReadString(' ')
	if e != nil {
		return 0, e
	}

	i, e := strconv.Atoi(strings.TrimSpace(tmp))
	if e != nil {
		return 0, e
	}

	return i, nil
}
