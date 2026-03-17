package orderrepo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	DB *pgxpool.Pool
}

type Good struct {
	Id         string  `json:"good_id"`
	NameOfGood string  `json:"name_of_good"`
	Price      float64 `json:"price"`
	Count      int64   `json:"count"`
}

type Order struct {
	Id         string  `json:"order_id"`
	UserId     string  `json:"user_id"`
	TotalPrice float64 `json:"total_price"`
	Status     string  `json:"status"`
}

type OrderItems struct {
	Id      string  `json:"id"`
	OrderId string  `json:"order_id"`
	GoodId  string  `json:"good_id"`
	Count   int64   `json:"count"`
	Price   float64 `json:"price"`
}

type GetOrderResponse struct {
	OrderInfo Order   `json:"order_info"`
	Goods     []*Good `json:"goods"`
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{
		DB: pool,
	}
}

func (r *Repo) CreateOrder(ctx context.Context, user_id string, total_price float64, goods []*Good) (string, error) {

	tx, err := r.DB.Begin(ctx)

	if err != nil {
		return "", err
	}

	for _, good := range goods {
		var count int64
		err := tx.QueryRow(ctx, "SELECT count FROM goods WHERE id=$1 FOR UPDATE", good.Id).Scan(&count)
		if err != nil {
			return "", err
		}

		if count < good.Count {
			return "", fmt.Errorf("count goods in storage much less than in order")
		}

		_, err = tx.Exec(ctx, "UPDATE goods SET count=count-$1 WHERE id=$2", good.Count, good.Id)
		if err != nil {
			return "", err
		}
	}

	defer tx.Rollback(ctx)

	var order_id string

	query := "INSERT INTO orders (user_id,total_price,status) VALUES($1,$2,$3) RETURNING id"

	err = tx.QueryRow(ctx, query, user_id, total_price, "CREATED").Scan(&order_id)

	if err != nil {
		return "", err
	}

	data := [][]any{}

	for _, item := range goods {
		data = append(data, []any{order_id, item.Id, item.NameOfGood, item.Count, item.Price})
	}

	_, err = tx.CopyFrom(ctx,
		pgx.Identifier{"order_items"},
		[]string{"order_id", "good_id", "name_of_good", "count", "price"},
		pgx.CopyFromRows(data),
	)

	if err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}

	return order_id, nil
}

func (r *Repo) GetOrderById(ctx context.Context, order_id string) (*GetOrderResponse, error) {

	query := `
	   SELECT o.id, o.user_id, o.total_price, o.status,
	   i.good_id,i.name_of_good,i.count, i.price
	   FROM orders o
	   INNER JOIN order_items i ON i.order_id=o.id
	   WHERE o.id = $1
	`

	rows, err := r.DB.Query(ctx, query, order_id)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var order GetOrderResponse
	no_rows := true

	for rows.Next() {
		var item Good
		err := rows.Scan(
			&order.OrderInfo.Id, &order.OrderInfo.UserId, &order.OrderInfo.TotalPrice, &order.OrderInfo.Status,
			&item.Id, &item.NameOfGood, &item.Count, &item.Price,
		)
		if err != nil {
			return nil, err
		}

		order.Goods = append(order.Goods, &item)

		no_rows = false
	}

	if no_rows {
		return nil, pgx.ErrNoRows
	}

	return &order, nil
}
