package postgres

import (
	"context"
	"log"
	"testing"

	"github.com/jmoiron/sqlx"
	db "github.com/mukhtarkv/workspace/kit/sql"
	"github.com/mukhtarkv/workspace/todo/todoapp"
	"github.com/mukhtarkv/workspace/todo/todoapp/assets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type todoTestSuite struct {
	suite.Suite
	st *storage
}

func TestToDoSuite(t *testing.T) {
	suite.Run(t, new(todoTestSuite))
}

func (s *todoTestSuite) SetupSuite() {
	log.Println("setup suite")

	rootdb, err := db.Open("postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		s.T().Fatalf("opening db connection %v", err)
	}
	s.st = &storage{
		db: rootdb,
	}

	// Apply migration scripts
	if err := db.Migrate(rootdb, "test_todo", assets.SF); err != nil {
		s.st.db.Close()
		s.T().Fatalf("applying schema migration scripts %v", err)
	}

	// Add test data
	if _, err := sqlx.LoadFile(rootdb, "./testdata/todo.sql"); err != nil {
		s.st.db.Close()
		s.T().Fatalf("loading test data %v", err)
	}
}

func (s *todoTestSuite) TearDownSuite() {
	defer s.st.db.Close()
	sqlx.LoadFile(s.st.db, "./testdata/todo_cleanup.sql")
}

func (s *todoTestSuite) TestFetch() {
	todo, err := s.st.Fetch(context.Background(), "xid1")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), todo)
	assert.Equal(s.T(), "xid1", todo.Id)
	assert.Equal(s.T(), "buy grocery", todo.Title)
	assert.Equal(s.T(), "buy milk and bread", todo.Details)
}

// func (s *todoTestSuite) TestList() {
// 	todo, err := s.st.List(context.Background())
// 	assert.NoError(s.T(), err)
// 	assert.NotNil(s.T(), todo)
// 	assert.Equal(s.T(), 2, len(todo))
// 	assert.Equal(s.T(), "xid2", todo[1].Id)
// 	assert.Equal(s.T(), "visit dentist", todo[1].Title)
// 	assert.Equal(s.T(), "get scaling", todo[1].Details)
// }

func (s *todoTestSuite) TestCreate() {
	ctx := context.Background()
	err := s.st.Create(ctx, todoapp.ToDoItem{
		Id: "xid3",
		Title: "Create Test",
		Details: "Create Test Details",
	})
	assert.NoError(s.T(), err)
	todo, err := s.st.Fetch(ctx, "xid3")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), todo)
	assert.Equal(s.T(), "xid3", todo.Id)
	assert.Equal(s.T(), "Create Test", todo.Title)
	assert.Equal(s.T(), "Create Test Details", todo.Details)
}

func (s *todoTestSuite) TestDelete() {
	ctx := context.Background()
	err := s.st.Create(ctx, todoapp.ToDoItem{
		Id: "xid4",
		Title: "Delete Test",
		Details: "Delete Test Details",
	})
	assert.NoError(s.T(), err)
	err = s.st.Delete(ctx, "xid4")
	assert.NoError(s.T(), err)
	_, err = s.st.Fetch(ctx, "xid4")
	assert.NotNil(s.T(), err)
}
