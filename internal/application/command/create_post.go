package command

import (
	"context"
	"go-blog-ddd/internal/adapter/e"
	"go-blog-ddd/internal/adapter/mdtohtml"
	"go-blog-ddd/internal/domain/aggregates"
	"go-blog-ddd/internal/domain/commands"
	"go-blog-ddd/internal/domain/values"
)

type CreatePostHandler struct {
	postRepo aggregates.PostRepository
}

func NewCreatePostHandler(repository aggregates.PostRepository) CreatePostHandler {
	if repository == nil {
		panic("missing PostRepository")
	}
	return CreatePostHandler{postRepo: repository}
}

func (c CreatePostHandler) Handle(ctx context.Context, cmd commands.CreatePost) (uint32, error) {
	uri, err := values.NewPostUri(cmd.Uri)
	if err != nil {
		return 0, e.PErrInvalidParam.WithError(err)
	}

	//if find, err := c.postRepo.FindByUri(ctx, uri); err != nil {
	//	return 0, e.RErrDatabase.WithError(err)
	//} else if find != nil {
	//	return 0, e.RErrResourceExists
	//}

	if err = c.postRepo.ResourceUniquenessCheck(ctx, uri); err != nil {
		return 0, err
	}

	htmlStr, err := mdtohtml.Convert(cmd.MDFile)
	if err != nil {
		return 0, err
	}

	nextID, err := c.postRepo.NextID(ctx)
	if err != nil {
		return 0, err
	}

	newPost := aggregates.NewPost(nextID, uri, htmlStr)

	if err = c.postRepo.Save(ctx, newPost); err != nil {
		return 0, err
	}
	return nextID, nil
}