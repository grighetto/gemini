package gemini

import (
	"errors"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/google/go-cmp/cmp"
)

type Session struct {
	testSession   *gocql.Session
	oracleSession *gocql.Session
}

var (
	ErrReadNoDataReturned = errors.New("read: no data returned")
)

func NewSession(testClusterHost string, oracleClusterHost string) *Session {
	testCluster := gocql.NewCluster(testClusterHost)
	testCluster.Timeout = 5 * time.Second
	testSession, err := testCluster.CreateSession()
	if err != nil {
		panic(err)
	}

	oracleCluster := gocql.NewCluster(oracleClusterHost)
	oracleCluster.Timeout = 5 * time.Second
	oracleSession, err := oracleCluster.CreateSession()
	if err != nil {
		panic(err)
	}

	return &Session{
		testSession:   testSession,
		oracleSession: oracleSession,
	}
}

func (s *Session) Close() {
	s.testSession.Close()
	s.oracleSession.Close()
}

func (s *Session) Mutate(query string, values ...interface{}) error {
	if err := s.testSession.Query(query, values...).Exec(); err != nil {
		return fmt.Errorf("%v [cluster = test, query = '%s']", err, query)
	}
	if err := s.oracleSession.Query(query, values...).Exec(); err != nil {
		return fmt.Errorf("%v [cluster = oracle, query = '%s']", err, query)
	}
	return nil
}

func (s *Session) Check(query string, values ...interface{}) error {
	testIter := s.testSession.Query(query, values...).Iter()
	oracleIter := s.oracleSession.Query(query, values...).Iter()
	for {
		testRow := make(map[string]interface{})
		if !testIter.MapScan(testRow) {
			break
		}
		oracleRow := make(map[string]interface{})
		if !oracleIter.MapScan(oracleRow) {
			break
		}
		diff := cmp.Diff(oracleRow, testRow)
		if diff != "" {
			return fmt.Errorf("rows differ (-%v +%v): %v", oracleRow, testRow, diff)
		}
		return nil
	}
	return ErrReadNoDataReturned
}
