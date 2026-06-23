package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"social/internal/store"
)

var usernames = []string{
	"alice", "bob", "charlie", "dave", "eve",
	"frank", "grace", "heidi", "ivan", "judy",
	"mallory", "oscar", "peggy", "trent", "victor",
	"walter", "sybil", "yvonne", "zack", "quinn",
	"nina", "leo", "mike", "sara", "tom",
	"uma", "vince", "wendy", "xavier", "yasmin",
	"zoe", "harry", "olivia", "liam", "noah",
	"emma", "ava", "sophia", "isabella", "mia",
	"lucas", "ethan", "james", "alex", "daniel",
	"chris", "jordan", "taylor", "casey", "riley",
}

var titles = []string{
	"Getting Started with Go",
	"Understanding REST APIs",
	"Building Your First Web App",
	"Introduction to Databases",
	"Clean Code Basics",
	"Debugging Made Easy",
	"Concurrency in Go",
	"Writing Better Functions",
	"API Design Tips",
	"Handling Errors Gracefully",
	"Working with JSON in Go",
	"Testing Your Code",
	"Structs and Interfaces",
	"Building a CLI Tool",
	"Go Modules Explained",
	"Optimizing Performance",
	"Logging Best Practices",
	"Middleware in Web Apps",
	"Intro to Microservices",
	"Deploying Your App",
}

var contents = []string{
	"Go is a simple and powerful language designed for building scalable systems.",
	"REST APIs allow communication between services using standard HTTP methods.",
	"Starting a web app requires routing, handlers, and a clear structure.",
	"Databases help store and retrieve data efficiently in applications.",
	"Clean code improves readability and makes maintenance easier.",
	"Debugging helps identify and fix issues quickly during development.",
	"Concurrency in Go is handled using goroutines and channels.",
	"Functions should be small, focused, and easy to understand.",
	"Good API design ensures clarity and ease of use for clients.",
	"Proper error handling prevents unexpected crashes in programs.",
	"JSON is commonly used for data exchange in web applications.",
	"Testing ensures your code works as expected and avoids regressions.",
	"Structs and interfaces are core building blocks in Go.",
	"CLI tools can automate repetitive tasks efficiently.",
	"Go modules manage dependencies and versioning بسهولة.",
	"Performance optimization improves speed and resource usage.",
	"Logging helps track application behavior and diagnose issues.",
	"Middleware adds reusable functionality to HTTP handlers.",
	"Microservices break applications into smaller, manageable services.",
	"Deployment makes your application accessible to users.",
}

var tags = []string{
	"go", "api", "web", "backend", "frontend",
	"database", "sql", "nosql", "devops", "cloud",
	"testing", "debugging", "performance", "security", "docker",
	"kubernetes", "microservices", "rest", "graphql", "cli",
}

var comments = []string{
	"Great post, really helpful!",
	"I learned something new today.",
	"This was easy to understand, thanks!",
	"Nice explanation, keep it up.",
	"Very insightful and practical.",
	"I’ve been looking for this, thanks!",
	"Clear and concise, well done.",
	"This helped me solve my issue.",
	"Awesome content, appreciate it!",
	"Good read, learned a lot.",
	"Simple and effective explanation.",
	"Thanks for sharing this.",
	"Really useful tips here.",
	"I like how you explained this.",
	"This makes things much clearer.",
	"Helpful for beginners like me.",
	"Great examples in this post.",
	"Looking forward to more posts.",
	"This was exactly what I needed.",
	"Well written and informative.",
}

func Seed(store store.Storage, db *sql.DB) {
	ctx := context.Background()
	tx, _ := db.BeginTx(ctx, nil)

	users := generateUsers(100)
	for _, user := range users {
		if err := store.Users.Create(ctx, tx, user); err != nil {
			_ = tx.Rollback()
			log.Println("Error seeding user", err)
			return
		}
	}

	posts := generatePosts(200, users)
	for _, post := range posts {
		if err := store.Posts.Create(ctx, post); err != nil {
			log.Println("Error seeding post", err)
			return
		}
	}

	comments := generateComments(500, users, posts)
	for _, comment := range comments {
		if err := store.Comments.Create(ctx, comment); err != nil {
			log.Println("Error seeding comment", err)
			return
		}
	}

	log.Println("Seeding completed...")
}

func generateUsers(num int) []*store.User {
	users := make([]*store.User, num)

	for i := 0; i < num; i++ {
		users[i] = &store.User{
			Username: usernames[i%len(usernames)] + fmt.Sprintf("%d", i),
			Email:    usernames[i%len(usernames)] + fmt.Sprintf("%d", i) + "@example.com",
			Role: store.Role{
				Name: "user",
			},
		}
	}
	return users
}

func generatePosts(num int, users []*store.User) []*store.Post {
	posts := make([]*store.Post, num)
	for i := 0; i < num; i++ {
		user := users[rand.Intn(len(users))]

		posts[i] = &store.Post{
			UserID:  user.ID,
			Title:   titles[rand.Intn(len(titles))],
			Content: contents[rand.Intn(len(contents))],
			Tags: []string{
				tags[rand.Intn(len(tags))],
				tags[rand.Intn(len(tags))],
			},
		}
	}
	return posts
}

func generateComments(num int, users []*store.User, posts []*store.Post) []*store.Comment {
	cms := make([]*store.Comment, num)
	for i := 0; i < num; i++ {
		cms[i] = &store.Comment{
			PostID:  posts[rand.Intn(len(posts))].ID,
			UserID:  users[rand.Intn(len(users))].ID,
			Content: comments[rand.Intn(len(comments))],
		}
	}
	return cms
}
