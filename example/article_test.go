package example

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

var authorBluePrint = func(i int, last Author) Author {
	num := i + 1
	return Author{
		ID:   num,
		Name: fmt.Sprintf("Author %d", num),
	}
}

var articleBluePrint = func(i int, last Article) Article {
	num := i + 1
	return Article{
		ID:                 num,
		Title:              fmt.Sprintf("Article %d", i+1),
		AuthorID:           num,
		PublishScheduledAt: time.Now().Add(-1 * time.Hour),
		PublishedAt:        time.Now().Add(-1 * time.Hour),
		LikeCount:          15,
	}
}

var articleTraitDraft = ArticleTrait{
	Article{
		Status: ArticleStatusDraft,
	},
	[]ArticleField{ArticlePublishedAtField},
}

// articleFactory.NewBuilder(articleBluePrint, []ArticleTrait{articleTraitDraft},  Article{Title: "OMG", LikeCount: 999}, articleTraitDraft).Zero(ArticlePublishedAtField).Values(Article{})

var articleTraitPublishScheduled = ArticleTrait{
	Article{
		Status:             ArticleStatusOpen,
		PublishScheduledAt: time.Now().Add(1 * time.Hour),
	},
	nil,
}

var articleTraitPublished = ArticleTrait{
	Article{
		Status:             ArticleStatusOpen,
		PublishScheduledAt: time.Now().Add(-1 * time.Hour),
		PublishedAt:        time.Now().Add(-1 * time.Hour),
		LikeCount:          15,
	},
	nil,
}

func TestArticleList_SelectPublished(t *testing.T) {
	articleFactory := NewArticleFactory(t)
	// if you want to persist articles, set OnBuild func here
	articleFactory.OnBuild(func(t *testing.T, article *Article) { fmt.Println("Insert to db here") })

	// creates 3 different articles
	waitReview, publishedScheduled, published := articleFactory.NewBuilder(articleBluePrint).
		EachParam(articleTraitDraft, articleTraitPublishScheduled, articleTraitPublished).
		Zero(ArticleLikeCountField).
		ResetAfter().
		Build3()

	tests := []struct {
		name string
		list ArticleList
		want ArticleList
	}{
		{
			name: "returns only published articles",
			list: ArticleList{waitReview, publishedScheduled, published},
			want: ArticleList{published},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.list.SelectPublished(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SelectPublished() = %v, want %v", got, tt.want)
			}
		})
	}

	if waitReview.Status != ArticleStatusDraft {
		t.Errorf("waitReview Article should be a Draft due to trait value, got %v", waitReview.Status)
	}
	if !waitReview.PublishedAt.IsZero() {
		t.Errorf("waitReview Article should have no PublishedAt due to trait's zero value, got %v", waitReview.PublishedAt)
	}
}

func TestArticleList_SelectAuthoredBy(t *testing.T) {
	authorFactory := NewAuthorFactory(t)
	articleFactory := NewArticleFactory(t)

	author1, author2 := authorFactory.NewBuilder(authorBluePrint).Build2()
	articlesAuthoredBy1 := articleFactory.NewBuilder(articleBluePrint).Set(Article{AuthorID: author1.ID}).BuildList(4)
	articleAuthoredBy2 := articleFactory.NewBuilder(articleBluePrint).Set(Article{AuthorID: author2.ID}).Build()

	type args struct {
		authorID int
	}
	tests := []struct {
		name string
		list ArticleList
		args args
		want ArticleList
	}{
		{
			name: "returns articles authored by author 1",
			list: append(articlesAuthoredBy1, articleAuthoredBy2),
			args: args{authorID: author1.ID},
			want: articlesAuthoredBy1,
		},
		{
			name: "returns articles authored by author 2",
			list: append(articlesAuthoredBy1, articleAuthoredBy2),
			args: args{authorID: author2.ID},
			want: ArticleList{articleAuthoredBy2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.list.SelectAuthoredBy(tt.args.authorID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SelectAuthoredBy() = %v, want %v", got, tt.want)
			}
		})
	}
}
