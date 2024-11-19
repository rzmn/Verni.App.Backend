package defaultRepository

import (
	"context"
	"errors"

	"github.com/rzmn/Verni.App.Backend/internal/db"
	"github.com/rzmn/Verni.App.Backend/internal/repositories"
	"github.com/rzmn/Verni.App.Backend/internal/repositories/spendings"
	"github.com/rzmn/Verni.App.Backend/internal/services/logging"

	"github.com/google/uuid"
)

func New(db db.DB, logger logging.Service) spendings.Repository {
	return &defaultRepository{
		db:     db,
		logger: logger,
	}
}

type defaultRepository struct {
	db     db.DB
	logger logging.Service
}

func (c *defaultRepository) AddExpense(expense spendings.Expense) repositories.MutationWorkItemWithReturnValue[spendings.ExpenseId] {
	const op = "repositories.spendings.postgresRepository.AddExpense"
	expenseId := spendings.ExpenseId(uuid.New().String())
	return repositories.MutationWorkItemWithReturnValue[spendings.ExpenseId]{
		Perform: func() (spendings.ExpenseId, error) {
			if err := c.addExpense(expense, expenseId); err != nil {
				c.logger.LogInfo("%s: failed to insert err: %v", op, err)
				return expenseId, err
			}
			return expenseId, nil
		},
		Rollback: func() error {
			return c.removeExpense(expenseId)
		},
	}
}

func (c *defaultRepository) addExpense(expense spendings.Expense, id spendings.ExpenseId) error {
	const op = "repositories.spendings.postgresRepository.addExpense"
	c.logger.LogInfo("%s: start[expense=%v id=%s]", op, expense, id)
	tx, err := c.db.BeginTx(context.Background(), nil)
	if err != nil {
		c.logger.LogInfo("%s: failed to create tx err: %v", op, err)
		return err
	}
	_, err = c.db.Exec(`
INSERT INTO 
	deals(id, timestamp, details, cost, currency) 
VALUES($1, $2, $3, $4, $5);
`, string(id), expense.Timestamp, expense.Details, int64(expense.Total), string(expense.Currency))
	if err != nil {
		c.logger.LogInfo("%s: failed to insert expense err: %v", op, err)
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
			c.logger.LogInfo("%s: failed to insert share %d err: %v", op, i, err)
			tx.Rollback()
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		c.logger.LogInfo("%s: failed to commit tx err: %v", op, err)
		return err
	}
	c.logger.LogInfo("%s: success[expense=%v id=%s]", op, expense, id)
	return nil
}

func (c *defaultRepository) RemoveExpense(expenseId spendings.ExpenseId) repositories.MutationWorkItem {
	const op = "repositories.spendings.postgresRepository.RemoveExpense"
	expense, err := c.GetExpense(expenseId)
	return repositories.MutationWorkItem{
		Perform: func() error {
			if err != nil {
				c.logger.LogInfo("%s: failed to get expense to remove err: %v", op, err)
				return err
			}
			if expense == nil {
				c.logger.LogInfo("%s: expense to remove not found", op)
				return errors.New("expense to remove not found")
			}
			return c.removeExpense(expenseId)
		},
		Rollback: func() error {
			if err != nil {
				c.logger.LogInfo("%s: failed to get expense to remove err: %v", op, err)
				return err
			}
			if expense == nil {
				c.logger.LogInfo("%s: expense to remove not found", op)
				return errors.New("expense to remove not found")
			}
			return c.addExpense(spendings.Expense((*expense).Expense), spendings.ExpenseId((*expense).Id))
		},
	}
}

func (c *defaultRepository) removeExpense(id spendings.ExpenseId) error {
	const op = "repositories.spendings.postgresRepository.removeExpense"
	c.logger.LogInfo("%s: start[id=%s]", op, id)
	tx, err := c.db.BeginTx(context.Background(), nil)
	if err != nil {
		c.logger.LogInfo("%s: failed to create tx err: %v", op, err)
		return err
	}
	_, err = c.db.Exec(`DELETE FROM deals WHERE id = $1;`, string(id))
	if err != nil {
		c.logger.LogInfo("%s: failed to remove expense err: %v", op, err)
		tx.Rollback()
		return err
	}
	_, err = c.db.Exec(`DELETE FROM spendings WHERE dealId = $1;`, string(id))
	if err != nil {
		c.logger.LogInfo("%s: failed to remove shares err: %v", op, err)
		tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		c.logger.LogInfo("%s: failed to commit tx err: %v", op, err)
		return err
	}
	c.logger.LogInfo("%s: success[id=%s]", op, id)
	return nil
}

func (c *defaultRepository) GetExpense(id spendings.ExpenseId) (*spendings.IdentifiableExpense, error) {
	const op = "repositories.spendings.postgresRepository.GetExpense"
	c.logger.LogInfo("%s: start[id=%s]", op, id)
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
		c.logger.LogInfo("%s: failed to perform query err: %v", op, err)
		return nil, err
	}
	defer rows.Close()
	var expense *spendings.IdentifiableExpense
	for rows.Next() {
		var _expense spendings.IdentifiableExpense
		if expense == nil {
			_expense = spendings.IdentifiableExpense{}
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
			c.logger.LogInfo("%s: scan failed err: %v", op, err)
			return nil, err
		}
		_expense.Id = spendings.ExpenseId(expenseIdString)
		_expense.Shares = append(_expense.Shares, spendings.ShareOfExpense{
			Counterparty: spendings.CounterpartyId(counterparty),
			Cost:         spendings.Cost(cost),
		})
		expense = &_expense
	}
	if err := rows.Err(); err != nil {
		c.logger.LogInfo("%s: found rows err: %v", op, err)
		return nil, err
	}
	c.logger.LogInfo("%s: success[id=%s]", op, id)
	return expense, nil
}

func (c *defaultRepository) GetExpensesBetween(counterparty1 spendings.CounterpartyId, counterparty2 spendings.CounterpartyId) ([]spendings.IdentifiableExpense, error) {
	const op = "repositories.spendings.postgresRepository.GetExpensesBetween"
	c.logger.LogInfo("%s: start[c1=%s c2=%s]", op, counterparty1, counterparty2)
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
		c.logger.LogInfo("%s: failed to perform query err: %v", op, err)
		return nil, err
	}
	defer rows.Close()
	expenses := []spendings.IdentifiableExpense{}
	for rows.Next() {
		expense := spendings.IdentifiableExpense{}
		expense.Shares = make([]spendings.ShareOfExpense, 2)
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
			c.logger.LogInfo("%s: scan failed err: %v", op, err)
			return nil, err
		}
		expense.Id = spendings.ExpenseId(expenseId)
		expense.Shares[0].Counterparty = spendings.CounterpartyId(uid1)
		expense.Shares[1].Counterparty = spendings.CounterpartyId(uid2)
		expenses = append(expenses, expense)
	}
	if err := rows.Err(); err != nil {
		c.logger.LogInfo("%s: found rows err: %v", op, err)
		return nil, err
	}
	c.logger.LogInfo("%s: success[c1=%s c2=%s]", op, counterparty1, counterparty2)
	return expenses, nil
}

func (c *defaultRepository) GetBalance(counterparty spendings.CounterpartyId) ([]spendings.Balance, error) {
	const op = "repositories.spendings.postgresRepository.GetBalance"
	c.logger.LogInfo("%s: start[counterparty=%s]", op, counterparty)
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
		c.logger.LogInfo("%s: failed to perform query err: %v", op, err)
		return nil, err
	}
	defer rows.Close()
	balancesMap := map[spendings.CounterpartyId]spendings.Balance{}
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
			c.logger.LogInfo("%s: scan failed err: %v", op, err)
			return []spendings.Balance{}, err
		}
		_, ok := balancesMap[spendings.CounterpartyId(user)]
		if !ok {
			balancesMap[spendings.CounterpartyId(user)] = spendings.Balance{
				Counterparty: spendings.CounterpartyId(user),
				Currencies:   map[spendings.Currency]spendings.Cost{},
			}
		}
		balancesMap[spendings.CounterpartyId(user)].Currencies[spendings.Currency(currency)] += spendings.Cost(cost)
	}
	if err := rows.Err(); err != nil {
		c.logger.LogInfo("%s: found rows err: %v", op, err)
		return nil, err
	}
	balance := make([]spendings.Balance, 0, len(balancesMap))
	if err != nil {
		c.logger.LogInfo("%s: unexpected err %v", op, err)
		return balance, err
	}
	for _, value := range balancesMap {
		balance = append(balance, value)
	}
	c.logger.LogInfo("%s: start[success=%s]", op, counterparty)
	return balance, nil
}
