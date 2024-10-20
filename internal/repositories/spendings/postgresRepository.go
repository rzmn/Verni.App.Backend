package spendings

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"verni/internal/repositories"
	"verni/internal/storage"

	"github.com/google/uuid"
)

type postgresRepository struct {
	db *sql.DB
}

func (c *postgresRepository) InsertDeal(deal Deal) repositories.MutationWorkItemWithReturnValue[DealId] {
	const op = "repositories.spendings.postgresRepository.InsertDeal"
	dealId := DealId(uuid.New().String())
	return repositories.MutationWorkItemWithReturnValue[DealId]{
		Perform: func() (DealId, error) {
			if err := c.insertDeal(deal, dealId); err != nil {
				log.Printf("%s: failed to insert err: %v", op, err)
				return dealId, err
			}
			return dealId, nil
		},
		Rollback: func() error {
			return c.removeDeal(dealId)
		},
	}
}

func (c *postgresRepository) insertDeal(deal Deal, did DealId) error {
	const op = "repositories.spendings.postgresRepository.insertDeal"
	log.Printf("%s: start[deal=%v did=%s]", op, deal, did)
	tx, err := c.db.BeginTx(context.Background(), nil)
	if err != nil {
		log.Printf("%s: failed to create tx err: %v", op, err)
		return err
	}
	_, err = c.db.Exec(`
INSERT INTO 
	deals(id, timestamp, details, cost, currency) 
VALUES($1, $2, $3, $4, $5);
`, string(did), deal.Timestamp, deal.Details, deal.Cost, deal.Currency)
	if err != nil {
		log.Printf("%s: failed to insert deal err: %v", op, err)
		tx.Rollback()
		return err
	}
	for i := 0; i < len(deal.Spendings); i++ {
		spending := deal.Spendings[i]
		_, err = c.db.Exec(`
INSERT INTO 
	spendings(id, dealId, cost, counterparty) 
VALUES($1, $2, $3, $4);
		`, uuid.New().String(), string(did), spending.Cost, string(spending.UserId))
		if err != nil {
			log.Printf("%s: failed to insert spending %d err: %v", op, i, err)
			tx.Rollback()
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		log.Printf("%s: failed to commit tx err: %v", op, err)
		return err
	}
	log.Printf("%s: success[deal=%v did=%s]", op, deal, did)
	return nil
}

func (c *postgresRepository) RemoveDeal(did DealId) repositories.MutationWorkItem {
	const op = "repositories.spendings.postgresRepository.RemoveDeal"
	deal, err := c.GetDeal(did)
	return repositories.MutationWorkItem{
		Perform: func() error {
			if err != nil {
				log.Printf("%s: failed to get deal to remove err: %v", op, err)
				return err
			}
			if deal == nil {
				log.Printf("%s: deal to remove not found", op)
				return errors.New("deal to remove not found")
			}
			return c.removeDeal(did)
		},
		Rollback: func() error {
			if err != nil {
				log.Printf("%s: failed to get deal to remove err: %v", op, err)
				return err
			}
			if deal == nil {
				log.Printf("%s: deal to remove not found", op)
				return errors.New("deal to remove not found")
			}
			return c.insertDeal(Deal((*deal).Deal), DealId((*deal).Id))
		},
	}
}

func (c *postgresRepository) removeDeal(did DealId) error {
	const op = "repositories.spendings.postgresRepository.removeDeal"
	log.Printf("%s: start[did=%s]", op, did)
	tx, err := c.db.BeginTx(context.Background(), nil)
	if err != nil {
		log.Printf("%s: failed to create tx err: %v", op, err)
		return err
	}
	_, err = c.db.Exec(`DELETE FROM deals WHERE id = $1;`, string(did))
	if err != nil {
		log.Printf("%s: failed to remove deal err: %v", op, err)
		tx.Rollback()
		return err
	}
	_, err = c.db.Exec(`DELETE FROM spendings WHERE dealId = $1;`, string(did))
	if err != nil {
		log.Printf("%s: failed to remove associated spendings err: %v", op, err)
		tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		log.Printf("%s: failed to commit tx err: %v", op, err)
		return err
	}
	log.Printf("%s: success[did=%s]", op, did)
	return nil
}

func (c *postgresRepository) GetDeal(did DealId) (*IdentifiableDeal, error) {
	const op = "repositories.spendings.postgresRepository.GetDeal"
	log.Printf("%s: start[did=%s]", op, did)
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
	rows, err := c.db.Query(query, string(did))
	if err != nil {
		log.Printf("%s: failed to perform query err: %v", op, err)
		return nil, err
	}
	defer rows.Close()
	var deal *IdentifiableDeal
	for rows.Next() {
		var _deal IdentifiableDeal
		if deal == nil {
			_deal = IdentifiableDeal{}
		} else {
			_deal = *deal
		}
		var dealIdString string
		var cost int64
		var counterparty string
		err = rows.Scan(
			&dealIdString,
			&_deal.Timestamp,
			&_deal.Details,
			&_deal.Cost,
			&_deal.Currency,
			&cost,
			&counterparty)
		if err != nil {
			log.Printf("%s: scan failed err: %v", op, err)
			return nil, err
		}
		_deal.Id = storage.DealId(dealIdString)
		_deal.Spendings = append(_deal.Spendings, storage.Spending{
			UserId: storage.UserId(counterparty),
			Cost:   cost,
		})
		deal = &_deal
	}
	if err := rows.Err(); err != nil {
		log.Printf("%s: found rows err: %v", op, err)
		return nil, err
	}
	log.Printf("%s: success[did=%s]", op, did)
	return deal, nil
}

func (c *postgresRepository) GetDeals(counterparty1 UserId, counterparty2 UserId) ([]IdentifiableDeal, error) {
	const op = "repositories.spendings.postgresRepository.GetDeals"
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
  s1.counterparty = $1 AND s2.counterparty = $2;
`
	rows, err := c.db.Query(query, string(counterparty1), string(counterparty2))
	if err != nil {
		log.Printf("%s: failed to perform query err: %v", op, err)
		return nil, err
	}
	defer rows.Close()
	deals := []IdentifiableDeal{}
	for rows.Next() {
		deal := IdentifiableDeal{}
		deal.Spendings = make([]storage.Spending, 2)
		var uid1 string
		var uid2 string
		var did string
		err = rows.Scan(
			&did,
			&uid1,
			&deal.Spendings[0].Cost,
			&uid2,
			&deal.Spendings[1].Cost,
			&deal.Timestamp,
			&deal.Details,
			&deal.Cost,
			&deal.Currency)
		if err != nil {
			log.Printf("%s: scan failed err: %v", op, err)
			return nil, err
		}
		deal.Id = storage.DealId(did)
		deal.Spendings[0].UserId = storage.UserId(uid1)
		deal.Spendings[1].UserId = storage.UserId(uid2)
		deals = append(deals, deal)
	}
	if err := rows.Err(); err != nil {
		log.Printf("%s: found rows err: %v", op, err)
		return nil, err
	}
	log.Printf("%s: success[c1=%s c2=%s]", op, counterparty1, counterparty2)
	return deals, nil
}

func (c *postgresRepository) GetCounterparties(uid UserId) ([]SpendingsPreview, error) {
	const op = "repositories.spendings.postgresRepository.GetCounterparties"
	log.Printf("%s: start[uid=%s]", op, uid)
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
	rows, err := c.db.Query(query, string(uid))
	if err != nil {
		log.Printf("%s: failed to perform query err: %v", op, err)
		return nil, err
	}
	defer rows.Close()
	spendingsMap := map[string]SpendingsPreview{}
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
			return []SpendingsPreview{}, err
		}
		_, ok := spendingsMap[user]
		if !ok {
			spendingsMap[user] = SpendingsPreview{
				Counterparty: user,
				Balance:      map[string]int64{},
			}
		}
		spendingsMap[user].Balance[currency] += cost
	}
	if err := rows.Err(); err != nil {
		log.Printf("%s: found rows err: %v", op, err)
		return nil, err
	}
	spendings := make([]SpendingsPreview, 0, len(spendingsMap))
	if err != nil {
		log.Printf("%s: unexpected err %v", op, err)
		return spendings, err
	}
	for _, value := range spendingsMap {
		spendings = append(spendings, value)
	}
	log.Printf("%s: success[uid=%s]", op, uid)
	return spendings, nil
}

func (c *postgresRepository) GetCounterpartiesForDeal(did DealId) ([]UserId, error) {
	const op = "repositories.spendings.postgresRepository.GetCounterpartiesForDeal"
	log.Printf("%s: start[did=%s]", op, did)
	query := `SELECT counterparty FROM spendings WHERE dealId = $1;`
	rows, err := c.db.Query(query, string(did))
	if err != nil {
		log.Printf("%s: failed to perform query err: %v", op, err)
		return nil, err
	}
	defer rows.Close()
	counterparties := []UserId{}
	for rows.Next() {
		var counterparty string
		if err := rows.Scan(&counterparty); err != nil {
			log.Printf("%s: scan failed err: %v", op, err)
			return []UserId{}, err
		}
		counterparties = append(counterparties, UserId(counterparty))
	}
	if err := rows.Err(); err != nil {
		log.Printf("%s: found rows err: %v", op, err)
		return nil, err
	}
	return counterparties, nil
}
