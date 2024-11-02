package spendings

import (
	"context"
	"errors"
	"log"
	"verni/internal/db"
	"verni/internal/repositories"

	"github.com/google/uuid"
)

type postgresRepository struct {
	db db.DB
}

func (c *postgresRepository) AddExpense(expense Expense) repositories.MutationWorkItemWithReturnValue[ExpenseId] {
	const op = "repositories.spendings.postgresRepository.AddExpense"
	expenseId := ExpenseId(uuid.New().String())
	return repositories.MutationWorkItemWithReturnValue[ExpenseId]{
		Perform: func() (ExpenseId, error) {
			if err := c.addExpense(expense, expenseId); err != nil {
				log.Printf("%s: failed to insert err: %v", op, err)
				return expenseId, err
			}
			return expenseId, nil
		},
		Rollback: func() error {
			return c.removeExpense(expenseId)
		},
	}
}

func (c *postgresRepository) addExpense(expense Expense, id ExpenseId) error {
	const op = "repositories.spendings.postgresRepository.addExpense"
	log.Printf("%s: start[expense=%v id=%s]", op, expense, id)
	tx, err := c.db.BeginTx(context.Background(), nil)
	if err != nil {
		log.Printf("%s: failed to create tx err: %v", op, err)
		return err
	}
	_, err = c.db.Exec(`
INSERT INTO 
	deals(id, timestamp, details, cost, currency) 
VALUES($1, $2, $3, $4, $5);
`, string(id), expense.Timestamp, expense.Details, int64(expense.Total), string(expense.Currency))
	if err != nil {
		log.Printf("%s: failed to insert expense err: %v", op, err)
		tx.Rollback()
		return err
	}
	for i := 0; i < len(expense.Shares); i++ {
		share := expense.Shares[i]
		_, err = c.db.Exec(`
INSERT INTO 
	spendings(id, dealId, cost, counterparty) 
VALUES($1, $2, $3, $4);
		`, uuid.New().String(), string(id), int64(share.Cost), string(share.Counterparty))
		if err != nil {
			log.Printf("%s: failed to insert share %d err: %v", op, i, err)
			tx.Rollback()
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		log.Printf("%s: failed to commit tx err: %v", op, err)
		return err
	}
	log.Printf("%s: success[expense=%v id=%s]", op, expense, id)
	return nil
}

func (c *postgresRepository) RemoveExpense(expenseId ExpenseId) repositories.MutationWorkItem {
	const op = "repositories.spendings.postgresRepository.RemoveExpense"
	expense, err := c.GetExpense(expenseId)
	return repositories.MutationWorkItem{
		Perform: func() error {
			if err != nil {
				log.Printf("%s: failed to get expense to remove err: %v", op, err)
				return err
			}
			if expense == nil {
				log.Printf("%s: expense to remove not found", op)
				return errors.New("expense to remove not found")
			}
			return c.removeExpense(expenseId)
		},
		Rollback: func() error {
			if err != nil {
				log.Printf("%s: failed to get expense to remove err: %v", op, err)
				return err
			}
			if expense == nil {
				log.Printf("%s: expense to remove not found", op)
				return errors.New("expense to remove not found")
			}
			return c.addExpense(Expense((*expense).Expense), ExpenseId((*expense).Id))
		},
	}
}

func (c *postgresRepository) removeExpense(id ExpenseId) error {
	const op = "repositories.spendings.postgresRepository.removeExpense"
	log.Printf("%s: start[id=%s]", op, id)
	tx, err := c.db.BeginTx(context.Background(), nil)
	if err != nil {
		log.Printf("%s: failed to create tx err: %v", op, err)
		return err
	}
	_, err = c.db.Exec(`DELETE FROM deals WHERE id = $1;`, string(id))
	if err != nil {
		log.Printf("%s: failed to remove expense err: %v", op, err)
		tx.Rollback()
		return err
	}
	_, err = c.db.Exec(`DELETE FROM spendings WHERE dealId = $1;`, string(id))
	if err != nil {
		log.Printf("%s: failed to remove shares err: %v", op, err)
		tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		log.Printf("%s: failed to commit tx err: %v", op, err)
		return err
	}
	log.Printf("%s: success[id=%s]", op, id)
	return nil
}

func (c *postgresRepository) GetExpense(id ExpenseId) (*IdentifiableExpense, error) {
	const op = "repositories.spendings.postgresRepository.GetExpense"
	log.Printf("%s: start[id=%s]", op, id)
	query := `
SELECT 
	d.id, 
	d.timestamp,
	d.details,
	d.cost,
	d.currency,
	s.cost,
	s.counterparty
FROM 
	deals d
	JOIN spendings s ON s.dealId = d.id
WHERE 
	d.id = $1;
`
	rows, err := c.db.Query(query, string(id))
	if err != nil {
		log.Printf("%s: failed to perform query err: %v", op, err)
		return nil, err
	}
	defer rows.Close()
	var expense *IdentifiableExpense
	for rows.Next() {
		var _expense IdentifiableExpense
		if expense == nil {
			_expense = IdentifiableExpense{}
		} else {
			_expense = *expense
		}
		var expenseIdString string
		var cost int64
		var counterparty string
		err = rows.Scan(
			&expenseIdString,
			&_expense.Timestamp,
			&_expense.Details,
			&_expense.Total,
			&_expense.Currency,
			&cost,
			&counterparty)
		if err != nil {
			log.Printf("%s: scan failed err: %v", op, err)
			return nil, err
		}
		_expense.Id = ExpenseId(expenseIdString)
		_expense.Shares = append(_expense.Shares, ShareOfExpense{
			Counterparty: CounterpartyId(counterparty),
			Cost:         Cost(cost),
		})
		expense = &_expense
	}
	if err := rows.Err(); err != nil {
		log.Printf("%s: found rows err: %v", op, err)
		return nil, err
	}
	log.Printf("%s: success[id=%s]", op, id)
	return expense, nil
}

func (c *postgresRepository) GetExpensesBetween(counterparty1 CounterpartyId, counterparty2 CounterpartyId) ([]IdentifiableExpense, error) {
	const op = "repositories.spendings.postgresRepository.GetExpensesBetween"
	log.Printf("%s: start[c1=%s c2=%s]", op, counterparty1, counterparty2)
	query := `
SELECT
  s1.dealId,
  s1.counterparty,
  s1.cost,
  s2.counterparty,
  s2.cost,
  d.timestamp,
  d.details,
  d.cost,
  d.currency
FROM
  spendings s1
  JOIN spendings s2 ON s1.dealId = s2.dealId
  JOIN deals d ON s1.dealId = d.id
WHERE
  s1.counterparty = $1 AND s2.counterparty = $2
ORDER BY d.timestamp;
`
	rows, err := c.db.Query(query, string(counterparty1), string(counterparty2))
	if err != nil {
		log.Printf("%s: failed to perform query err: %v", op, err)
		return nil, err
	}
	defer rows.Close()
	expenses := []IdentifiableExpense{}
	for rows.Next() {
		expense := IdentifiableExpense{}
		expense.Shares = make([]ShareOfExpense, 2)
		var uid1 string
		var uid2 string
		var expenseId string
		err = rows.Scan(
			&expenseId,
			&uid1,
			&expense.Shares[0].Cost,
			&uid2,
			&expense.Shares[1].Cost,
			&expense.Timestamp,
			&expense.Details,
			&expense.Total,
			&expense.Currency)
		if err != nil {
			log.Printf("%s: scan failed err: %v", op, err)
			return nil, err
		}
		expense.Id = ExpenseId(expenseId)
		expense.Shares[0].Counterparty = CounterpartyId(uid1)
		expense.Shares[1].Counterparty = CounterpartyId(uid2)
		expenses = append(expenses, expense)
	}
	if err := rows.Err(); err != nil {
		log.Printf("%s: found rows err: %v", op, err)
		return nil, err
	}
	log.Printf("%s: success[c1=%s c2=%s]", op, counterparty1, counterparty2)
	return expenses, nil
}

func (c *postgresRepository) GetBalance(counterparty CounterpartyId) ([]Balance, error) {
	const op = "repositories.spendings.postgresRepository.GetBalance"
	log.Printf("%s: start[counterparty=%s]", op, counterparty)
	query := `
SELECT
  d.currency,
  s2.counterparty,
  s1.cost
FROM
  deals d
  JOIN spendings s1 ON s1.dealId = d.id
  JOIN spendings s2 ON s2.dealId = d.id
WHERE 
  s1.counterparty = $1 AND s2.counterparty != $1;
`
	rows, err := c.db.Query(query, string(counterparty))
	if err != nil {
		log.Printf("%s: failed to perform query err: %v", op, err)
		return nil, err
	}
	defer rows.Close()
	balancesMap := map[CounterpartyId]Balance{}
	for rows.Next() {
		var currency string
		var user string
		var cost int64
		err = rows.Scan(
			&currency,
			&user,
			&cost,
		)
		if err != nil {
			log.Printf("%s: scan failed err: %v", op, err)
			return []Balance{}, err
		}
		_, ok := balancesMap[CounterpartyId(user)]
		if !ok {
			balancesMap[CounterpartyId(user)] = Balance{
				Counterparty: CounterpartyId(user),
				Currencies:   map[Currency]Cost{},
			}
		}
		balancesMap[CounterpartyId(user)].Currencies[Currency(currency)] += Cost(cost)
	}
	if err := rows.Err(); err != nil {
		log.Printf("%s: found rows err: %v", op, err)
		return nil, err
	}
	balance := make([]Balance, 0, len(balancesMap))
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return balance, err
	}
	for _, value := range balancesMap {
		balance = append(balance, value)
	}
	log.Printf("%s: start[success=%s]", op, counterparty)
	return balance, nil
}
