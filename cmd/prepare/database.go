package main

import (
	"github.com/rzmn/governi/internal/db"
	"github.com/rzmn/governi/internal/services/logging"
)

type databaseActions struct {
	setup func()
	drop  func()
}

func createDatabaseActions(database db.DB, logger logging.Service) databaseActions {
	return databaseActions{
		setup: func() {
			for _, table := range tables() {
				if err := table.create(database); err != nil {
					logger.LogInfo("failed to create table %s err: %v", err, table.name)
				}
				logger.LogInfo("created table %s", table.name)
			}
		},
		drop: func() {
			for _, table := range tables() {
				if err := table.delete(database); err != nil {
					logger.LogInfo("failed to drop table %s err: %v", err, table.name)
				}
				logger.LogInfo("droped table %s", table.name)
			}
		},
	}
}

type table struct {
	name   string
	create func(db db.DB) error
	delete func(db db.DB) error
}

func tables() []table {
	return []table{
		{
			name: "credentials",
			create: func(db db.DB) error {
				_, err := db.Exec(`
				CREATE TABLE credentials(
					id text NOT NULL PRIMARY KEY, 
					email text NOT NULL, 
					password text NOT NULL, 
					token text NOT NULL, 
					emailVerified bool NOT NULL
				);`)
				return err
			},
			delete: func(db db.DB) error {
				_, err := db.Exec(`DROP TABLE credentials;`)
				return err
			},
		},
		{
			name: "users",
			create: func(db db.DB) error {
				_, err := db.Exec(`
				CREATE TABLE users(
					id text NOT NULL PRIMARY KEY, 
					displayName text NOT NULL,
					avatarId text
				);`)
				return err
			},
			delete: func(db db.DB) error {
				_, err := db.Exec(`DROP TABLE users;`)
				return err
			},
		},
		{
			name: "friendRequests",
			create: func(db db.DB) error {
				_, err := db.Exec(`
				CREATE TABLE friendRequests(
					sender text NOT NULL, 
					target text NOT NULL,
					PRIMARY KEY(sender, target)
				);`)
				return err
			},
			delete: func(db db.DB) error {
				_, err := db.Exec(`DROP TABLE friendRequests;`)
				return err
			},
		},
		{
			name: "spendings",
			create: func(db db.DB) error {
				_, err := db.Exec(`
				CREATE TABLE spendings(
					id text NOT NULL PRIMARY KEY, 
					dealId text NOT NULL, 
					cost int NOT NULL, 
					counterparty text NOT NULL
				);`)
				return err
			},
			delete: func(db db.DB) error {
				_, err := db.Exec(`DROP TABLE spendings;`)
				return err
			},
		},
		{
			name: "deals",
			create: func(db db.DB) error {
				_, err := db.Exec(`
				CREATE TABLE deals(
					id text NOT NULL PRIMARY KEY, 
					timestamp int NOT NULL, 
					details text NOT NULL, 
					cost int NOT NULL, 
					currency text NOT NULL
				);`)
				return err
			},
			delete: func(db db.DB) error {
				_, err := db.Exec(`DROP TABLE deals;`)
				return err
			},
		},
		{
			name: "images",
			create: func(db db.DB) error {
				_, err := db.Exec(`
				CREATE TABLE images(
					id text NOT NULL PRIMARY KEY, 
					base64 text NOT NULL
				);`)
				return err
			},
			delete: func(db db.DB) error {
				_, err := db.Exec(`DROP TABLE images;`)
				return err
			},
		},
		{
			name: "pushTokens",
			create: func(db db.DB) error {
				_, err := db.Exec(`
				CREATE TABLE pushTokens(
					id text NOT NULL PRIMARY KEY, 
					token text NOT NULL
				);`)
				return err
			},
			delete: func(db db.DB) error {
				_, err := db.Exec(`DROP TABLE pushTokens;`)
				return err
			},
		},
		{
			name: "emailVerification",
			create: func(db db.DB) error {
				_, err := db.Exec(`
				CREATE TABLE emailVerification(
					email text NOT NULL PRIMARY KEY, 
					code text
				);`)
				return err
			},
			delete: func(db db.DB) error {
				_, err := db.Exec(`DROP TABLE emailVerification;`)
				return err
			},
		},
	}
}
