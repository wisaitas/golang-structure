package gormx

import (
	"context"

	"gorm.io/gorm"
)

type BaseRepository[T any] interface {
	Find(ctx context.Context, items *[]T, pagination *PaginationQuery, condition *Condition, relations *[]Relation) error
	FindOne(ctx context.Context, item *T, condition *Condition, relations *[]Relation) error
	Create(ctx context.Context, value any) error
	Update(ctx context.Context, value any) error
	Save(ctx context.Context, value any) error
	Delete(ctx context.Context, value any) error
	Transaction(ctx context.Context, fn func(txRepo BaseRepository[T]) error) error
	WithTx(tx *gorm.DB) BaseRepository[T]
	GetDB(ctx context.Context) *gorm.DB
}

func NewBaseRepository[T any](db *gorm.DB) BaseRepository[T] {
	return &baseRepository[T]{
		db: db,
	}
}

type baseRepository[T any] struct {
	db *gorm.DB
}

func (r *baseRepository[T]) Find(ctx context.Context, items *[]T, pagination *PaginationQuery, condition *Condition, relations *[]Relation) error {
	query := r.db.WithContext(ctx)

	if condition != nil {
		query = query.Where(condition.Query, condition.Args...)
	}

	if pagination != nil {
		if pagination.Page != nil && pagination.PageSize != nil {
			offset := *pagination.Page * *pagination.PageSize
			query = query.Offset(offset).Limit(*pagination.PageSize)
		}

		if pagination.Sort != nil && pagination.Order != nil {
			orderClause := *pagination.Sort + " " + *pagination.Order
			query = query.Order(orderClause)
		}
	}

	if relations != nil {
		for _, relation := range *relations {
			query = query.Preload(relation.Query, relation.Args...)
		}
	}

	return query.Find(items).Error
}

func (r *baseRepository[T]) FindOne(ctx context.Context, item *T, condition *Condition, relations *[]Relation) error {
	query := r.db.WithContext(ctx)

	if condition != nil {
		query = query.Where(condition.Query, condition.Args...)
	}

	if relations != nil {
		for _, relation := range *relations {
			query = query.Preload(relation.Query, relation.Args...)
		}
	}

	return query.First(item).Error
}

func (r *baseRepository[T]) Create(ctx context.Context, value any) error {
	return r.db.WithContext(ctx).Create(value).Error
}

func (r *baseRepository[T]) Update(ctx context.Context, value any) error {
	return r.db.WithContext(ctx).Updates(value).Error
}

func (r *baseRepository[T]) Save(ctx context.Context, value any) error {
	return r.db.WithContext(ctx).Save(value).Error
}

func (r *baseRepository[T]) Delete(ctx context.Context, value any) error {
	return r.db.WithContext(ctx).Delete(value).Error
}

func (r *baseRepository[T]) Transaction(ctx context.Context, fn func(txRepo BaseRepository[T]) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(r.WithTx(tx))
	})
}

func (r *baseRepository[T]) WithTx(tx *gorm.DB) BaseRepository[T] {
	return &baseRepository[T]{
		db: tx,
	}
}

func (r *baseRepository[T]) GetDB(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}
