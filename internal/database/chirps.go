package database

type Chirp struct {
	ID       int    `json:"id"`
	Body     string `json:"body"`
	AuthorID int    `json:"author_id"`
}

func (db *DB) CreateChirp(post string, authorID int) (Chirp, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	id := len(dbStruct.Chirps) + 1
	chirp := Chirp{
		ID:       id,
		Body:     post,
		AuthorID: authorID,
	}
	dbStruct.Chirps[id] = chirp

	err = db.writeDB(dbStruct)
	if err != nil {
		return Chirp{}, err
	}
	return chirp, nil
}

func (db *DB) GetChirps() ([]Chirp, error) {
	dbBuf, err := db.loadDB()
	if err != nil {
		return []Chirp{}, err
	}
	data := make([]Chirp, 0, len(dbBuf.Chirps))

	//_ := json.Unmarshal(dbBuf.Chirps, &data)
	for _, v := range dbBuf.Chirps {
		data = append(data, v)
	}
	return data, nil
}

func (db *DB) GetChirpByID(ID int) (Chirp, error) {
	dbBuf, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}
	for _, v := range dbBuf.Chirps {
		if v.ID == ID {
			return v, nil
		}
	}
	return Chirp{}, ErrNotExist
}

func (db *DB) DeleteChirp(ID int, authorID int) (Chirp, error) {
	dbBuf, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}
	for i, v := range dbBuf.Chirps {
		if v.ID == ID && v.AuthorID == authorID {
			deletedChirp := dbBuf.Chirps[i]
			dbBuf.Chirps[i] = Chirp{}
			return deletedChirp, nil
		}
	}
	return Chirp{}, ErrNotExist
}
