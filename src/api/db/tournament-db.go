package db

import (
	"strings"
)

const (
	announceTournamentQuery      = "insert into Tournaments values (?, ?, '')"
	selectPlayersTournamentQuery = "select Players from Tournaments where TourId=?"
	selectTournamentQuery        = "select * from Tournaments where TourId=?"
	updateTournamentPlayersQuery = "update Tournaments set Players=? where TourId = ?"
)

type Tournament struct {
	Id      int
	Deposit int
	Players []string
}

func (d *Db) CreateTournament(id int, deposit int) (rerr error) {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if rerr != nil {
			tx.Rollback()
		}
	}()

	stmt, err := tx.Prepare(announceTournamentQuery)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(id, deposit)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (d *Db) TournamentInfo(tourId int) (_ *Tournament, rerr error) {
	tx, err := d.db.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		if rerr != nil {
			tx.Rollback()
		}
	}()

	rows, err := tx.Query(selectTournamentQuery, tourId)
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		return nil, ErrorNotFound
	}

	var id, depo int
	var p string
	if err := rows.Scan(&id, &depo, &p); err != nil {
		return nil, err
	}

	defer tx.Commit()

	players := []string{}
	if p != "" {
		players = strings.Split(p, ",")
	}
	return &Tournament{id, depo, players}, nil
}

func (d *Db) JoinTournament(tourId int, playerId string) (rerr error) {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if rerr != nil {
			tx.Rollback()
		}
	}()

	rows, err := tx.Query(selectPlayersTournamentQuery, tourId)
	if err != nil {
		return err
	}

	if !rows.Next() {
		return ErrorNotFound
	}

	var players string
	if err := rows.Scan(&players); err != nil {
		return err
	}

	pArr := strings.Split(players, ",")
	for _, p := range pArr {
		if p == playerId {
			return ErrAlreadyExists
		}
	}

	semicolonRequired := len(pArr) >= 1 && pArr[0] != ""
	if semicolonRequired {
		players += "," + playerId
	} else {
		players += playerId
	}

	stmt, err := tx.Prepare(updateTournamentPlayersQuery)
	if err != nil {
		return err
	}
	if _, err = stmt.Exec(players, tourId); err != nil {
		return err
	}

	return tx.Commit()
}
