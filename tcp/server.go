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

func (s *Server) process(conn net.Conn) {
	defer conn.Close()

	r := bufio.NewReader(conn)
	for { // for 循环的目的就是当一切都没出错的话, 不用退出(当然也不用关闭 conn), 继续等待下一次的操作
		op, err := r.ReadByte() // 阻塞式等待
		if err != nil {
			// 不是读完了, 而是出错了, 记录日志
			if err != io.EOF {
				log.Println("close connection due to error:", err)
				return
			}
			// 读完了(实际上应该是什么都没读到), 就直接返回(这是退出条件)
			return
		}

		// 后续处理也有可能出错
		if op == 'S' {
			err = s.set(conn, r)
		} else if op == 'G' {
			err = s.get(conn, r)
		} else if op == 'D' {
			err = s.del(conn, r)
		} else {
			// 后续处理本身也有可能出错
			log.Println("close connection due to invalid operation:", err)
			return
		}

		// 出错就直接返回
		if err != nil {
			log.Println("close connection due to error:", err)
			return
		}

		// 就是说只要有无效操作就直接断开当前客户端与服务器的连接并返回
		// 读完了就直接返回
	}
}

func (s *Server) set(conn net.Conn, r *bufio.Reader) error {
	k, v, e := s.readKeyAndValue(r)
	if e != nil {
		return e
	}

	return sendResponse(nil, s.c.Set(k, v), conn)
}

func (s *Server) get(conn net.Conn, r *bufio.Reader) error {
	k, e := s.readKey(r)
	if e != nil {
		return e
	}

	v, e := s.c.Get(k)
	return sendResponse(v, e, conn)
}

func (s *Server) del(conn net.Conn, r *bufio.Reader) error {
	k, e := s.readKey(r)
	if e != nil {
		return e
	}

	return sendResponse(nil, s.c.Del(k), conn)
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
	_, e := conn.Write(append([]byte(vlen), value...)) // string 通过 ... 这样的操作可以转成字节数组?
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
