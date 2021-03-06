package presentation

import (
	"fmt"
	"log"

	a "github.com/HirokInoue/realtimeweb/application"
	i "github.com/HirokInoue/realtimeweb/infra"
)

const (
	listenStop = iota
)

type Handler interface {
	exec(*Client, interface{})
}

func NewAddCommentHandler() (*AddCommentHandler, error) {
	session, err := i.NewSession("realtimeweb")
	if err != nil {
		return nil, err
	}
	repos := i.NewCommentsRepository(session)
	s := a.NewCommentService(repos)
	return &AddCommentHandler{
		service: s,
	}, nil
}

type AddCommentHandler struct {
	service *a.CommentService
}

func (ah *AddCommentHandler) exec(c *Client, data interface{}) {
	isOk := true
	err := ah.service.Add(fmt.Sprintf("%s", data))
	if err != nil {
		log.Print(err)
		isOk = false
	}
	c.send <- Body{Name: "add comment", Ok: isOk}
}

func NewListenCommentsHandler() (*ListenCommentsHandler, error) {
	// FIXME: Don't Repeat Yourself!
	session, err := i.NewSession("realtimeweb")
	if err != nil {
		return nil, err
	}
	repos := i.NewCommentsRepository(session)
	s := a.NewCommentService(repos)
	return &ListenCommentsHandler{
		service: s,
	}, nil
}

type ListenCommentsHandler struct {
	service *a.CommentService
}

func (lh *ListenCommentsHandler) exec(c *Client, data interface{}) {
	ctx := c.NewStopContext(listenStop)
	go func() {
		str := make(chan string)
		err := make(chan error)

		lh.service.Listen(str, err, ctx)
		body := Body{
			Name: "listen comments",
		}
		for {
			select {
			case content := <-str:
				body.Ok = true
				body.Data = content
			case e := <-err:
				body.Ok = false
				body.Data = ""
				log.Println(e)
			case <-ctx.Done():
				return
			}
			c.send <- body
		}
	}()
}
