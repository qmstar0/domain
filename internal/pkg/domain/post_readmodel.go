package domain

import (
	"context"
	"github.com/qmstar0/nightsky-api/internal/apps/query"
	"github.com/qmstar0/nightsky-api/internal/pkg/e"
	"github.com/qmstar0/nightsky-api/internal/pkg/utils"
	"github.com/qmstar0/nightsky-api/pkg/transaction"
	"github.com/uptrace/bun"
)

type PostReadModel struct {
	db transaction.TransactionContext
}

func NewPostReadModel(db transaction.TransactionContext) PostReadModel {
	return PostReadModel{db: db}
}

func (p PostReadModel) FindByID(ctx context.Context, pid uint32) (query.PostView, error) {
	var posts = []*PostM{{ID: pid, Visible: true}}
	err := p.db.NewSelect().
		Model(&posts).
		Relation("PostTags").
		Relation("Category").
		WherePK("id", "visible").
		Scan(ctx)
	if err != nil {
		return query.PostView{}, e.RErrDatabase.WithError(err)
	}
	if len(posts) <= 0 {
		return query.PostView{}, e.RErrResourceNotExists
	}
	return PostModelToView(posts[0]), nil
}

func (p PostReadModel) FindByUri(ctx context.Context, uri string) (query.PostView, error) {
	var posts = []*PostM{{Uri: uri, Visible: true}}

	err := p.db.NewSelect().
		Model(&posts).
		Relation("PostTags").
		Relation("Category").
		WherePK("uri", "visible").
		Limit(1).
		Scan(ctx)
	if err != nil {
		return query.PostView{}, e.RErrDatabase.WithError(err)
	}
	if len(posts) <= 0 {
		return query.PostView{}, e.RErrResourceNotExists
	}
	return PostModelToView(posts[0]), nil
}

func (p PostReadModel) RecentPosts(ctx context.Context, limit int) (query.PostListView, error) {
	var posts = make([]*PostM, 0)

	err := p.db.NewSelect().
		Model(&posts).
		ColumnExpr("id, uri, title, created_at, updated_at").
		Order("created_at", "updated_at").
		Where("post.visible = ?", true).
		Limit(limit).
		Scan(ctx)
	if err != nil {
		return query.PostListView{}, e.RErrDatabase.WithError(err)
	}
	return PostModelToListView(posts), nil
}

func (p PostReadModel) AllWithFilter(
	ctx context.Context,
	limit,
	page int,
	tags []string,
	categroyID uint32,
	onlyVisible bool,
) (query.PostListView, error) {

	var posts = make([]*PostM, 0)

	selectQuery := p.db.NewSelect()
	totalQuery := p.db.NewSelect()
	//计算偏移
	offset := utils.Offset(page, limit)

	//category filter
	var CategoryFilterFn func(*bun.SelectQuery) *bun.SelectQuery
	if categroyID != 0 {
		CategoryFilterFn = func(selectQuery *bun.SelectQuery) *bun.SelectQuery {
			return selectQuery.Where("category.id = ?", categroyID)
		}
	}

	//main query
	selectQuery = selectQuery.
		Model(&posts).
		ExcludeColumn("content").
		Relation("PostTags").
		Relation("Category", CategoryFilterFn).
		Order("created_at DESC")

	totalQuery = totalQuery.
		Model((*PostM)(nil)).
		ColumnExpr("COUNT(*) as total")

	if onlyVisible {
		selectQuery = selectQuery.Where("visible = ?", true)
		totalQuery = totalQuery.Where("visible = ?", true)
	}
	if tagsLen := len(tags); tagsLen != 0 {

		HasQueryTag := p.db.NewSelect().
			Model((*PostTagM)(nil)).
			ColumnExpr("post_id").
			Where("tag IN (?)", bun.In(tags)).
			Group("post_id").
			Having("COUNT(*) >= ?", tagsLen)

		selectQuery = selectQuery.Where("post.id IN (?)", HasQueryTag)
		totalQuery = totalQuery.Where("post.id IN (?)", HasQueryTag)
	}
	if limit > 0 {
		selectQuery = selectQuery.Limit(limit)
	}
	if offset > 0 {
		selectQuery = selectQuery.Offset(offset)
	}

	if err := selectQuery.Scan(ctx); err != nil {
		return query.PostListView{}, e.RErrDatabase.WithError(err)
	}
	var total int
	if err := totalQuery.Scan(ctx, &total); err != nil {
		return query.PostListView{}, e.RErrDatabase.WithError(err)
	}

	result := PostModelToListView(posts)
	result.Page = page
	result.Prev, result.Next = utils.Paginator(total, page, limit)
	return result, nil
}
