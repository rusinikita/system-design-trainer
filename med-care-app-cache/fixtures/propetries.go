package main

import (
	"iter"
	"math/rand/v2"
	"slices"
)

type FixtureProperties struct {
	UsersCount         int
	ArticlesCount      int
	SegmentTypesCount  int
	MinUserSegments    int
	MaxUserSegments    int
	MinArticleSegments int
	MaxArticleSegments int
	MinSteps           int
	MaxSteps           int
}

func DefaultFixtureProperties() FixtureProperties {
	return FixtureProperties{
		UsersCount:         1_000_000,
		ArticlesCount:      10_000,
		SegmentTypesCount:  50,
		MinUserSegments:    2,
		MaxUserSegments:    7,
		MinArticleSegments: 1,
		MaxArticleSegments: 5,
		MinSteps:           5,
		MaxSteps:           50,
	}
}

// что делает эмуляция
// 1 - постоянный рандомный поток запросов (уведомление о событии из плана или просто заход по своей причине)
// 2 - раз в "день" уведомление о новой статье все в сегменте, 20% открывают пуш

func (p *FixtureProperties) NextRandomArticleSegmentsCount(r *rand.Rand) int {
	return r.IntN(p.MaxArticleSegments-p.MinArticleSegments) + p.MinArticleSegments
}

func (p *FixtureProperties) NextRandomPlanStepsCount(r *rand.Rand) int {
	return r.IntN(p.MaxSteps-p.MinSteps) + p.MinSteps
}

func (p *FixtureProperties) NextRandomUserID(r *rand.Rand) int64 {
	return int64(r.IntN(p.UsersCount))
}

func (p *FixtureProperties) NextRandomSegmentUsers(r *rand.Rand) iter.Seq[int64] {
	segmentID := p.nextRandomSegmentID(r)

	return p.randomUsersForSegment(segmentID)
}

func (p *FixtureProperties) UserSegments(id int64) []int64 {
	segmentsCount := int(id)%(p.MaxUserSegments-p.MinUserSegments) + p.MinUserSegments

	result := make([]int64, segmentsCount)
	for i := 0; i < segmentsCount; i++ {
		result[i] = id % int64(i*10)
	}

	return slices.Compact(result)
}

func (p *FixtureProperties) randomUsersForSegment(segmentID int64) iter.Seq[int64] {
	percent := rand.IntN(10) + 10
	usersToScan := p.UsersCount / 100 * percent
	scanStart := rand.IntN(p.UsersCount / 2)

	return func(yield func(int64) bool) {
		for userID := scanStart; userID < usersToScan; userID++ {
			segmentsCount := userID%(p.MaxUserSegments-p.MinUserSegments) + p.MinUserSegments

			for i := 0; i < segmentsCount; i++ {
				if segmentID != int64(userID%(i*10)) {
					continue
				}

				if !yield(int64(userID)) {
					return
				}
			}
		}
	}
}

func (p *FixtureProperties) nextRandomSegmentID(r *rand.Rand) int64 {
	return int64(r.IntN(p.SegmentTypesCount))
}
