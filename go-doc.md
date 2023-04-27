# Coding Notes for Golang



**Golang** is a powerful programming language that offers a number of features and functionalities that make it a popular choice for many developers. However, in order to write high-quality and maintainable code in Go, it is important to follow best practices and adhere to clean code principles.

In this document, we will explore these topics in more details. For every Golang programmerr, understanding these best practices and concepts can help him to write more efficient, maintainable, and scalable **SOFT**ware.


## Golang Best practices 

- Use ``gofmt`` to format your code: Golang has a built-in tool called ``gofmt`` that formats your code to adhere to the language's style guide. Consistently formatted code makes it easier to read and maintain.

- Avoid global variables: Global variables can lead to unexpected behavior and make code difficult to reason about. Instead, use function parameters and return values to pass data between functions.

- Use interfaces to define behavior: Interfaces provide a way to define behavior without specifying implementation details. By using interfaces, you can write more flexible and reusable code. PS: The single method interfaces are recommended in Golang and frequently used: add  `er` the method's name to get the interface's name.  
**Example**:

  ```go
  type Writer interface {
          Write(p []byte) (n int, err error)
  }
  ```
- Use ``goroutines`` and ``channels`` for **concurrency**: Golang provides built-in support for concurrency through goroutines and channels. By using these features, you can write efficient and scalable code that can run multiple functions concurrently. (see example in the next session)

- Use error handling: Golang provides a built-in error handling mechanism that makes it easy to handle errors gracefully. Always handle errors in your code to prevent unexpected behavior. ALso DO NEVER return void (void does not exist in Golang anyways), EVERY function should return a value, at least an error ( that can be ``nil``)

- Use ``defer`` to clean up resources: Golang provides a defer statement that allows you to schedule a function call to be executed after the current function returns. Use defer to release resources such as files, database connections, or locks.
It's recommanded to use defer right after allocating the resource.    
**Example:**
    ```go
    // imports

    func main() {
        db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/mydatabase")
        if err != nil {
            panic(err.Error())
        }
      // defer after making sure that the db is not nil, i.e the error is nil
        defer db.Close() 

        // some operations...
    }
    ```

- Write **tests**: Golang has a built-in testing package that makes it easy to write tests for your code. Writing tests helps to ensure that your code works as intended and makes it easier to maintain and refactor in the future.


## Goroutines and Channels   

`Goroutines` and `channels` are powerful features in Golang that allow you to write efficient and scalable software that can run multiple functions concurrently within the same address space.

``goroutines`` are lightweight threads that enable you to run multiple functions concurrently. You can create a goroutine by using the `go` keyword followed by the function you want to execute concurrently. Here's an example of a goroutine that registers a user:

```go
type User struct {
    Name string
    Email string
    Password string
}

func registerUser(name string, email string, ,password string, result chan User) {
    user := User{
        Name: name,
        Email: email,
        Password:password
    }
    result <- user
}

func main() {
    result := make(chan User) 
    go registerUser("John Doe", "john@example.com", "123456$$", result)
    user := <- result
    fmt.Println("User registred successfully:", user)
}`
```
In this example, the `registerUser()` function creates a ``User`` struct with a given name and age, and sends it through the result channel. The `main()` function receives the `user` from the channel and prints it to the console.

PS: there are two types of channels in Golang: **buffered** and **unbuffered** channels. CHannels have their mechanism to handle errors and receive data, selct

**Channels** are used to communicate between **goroutines**, allowing them to synchronize their execution and exchange data. You can create a **channel** by using the make function followed by the type of data you want to send and receive through the channel.

By using **goroutines** and **channels**, you can create efficient and scalable software that can run multiple functions concurrently and communicate between them seamlessly.

 **Note**   

  If you want to perform non-blocking operations on a channel, you can use ``select`` with a default clause. This will allow your program to continue executing if the channel operation would block.


## Clean code principals ( by Uncle Bob )

- Keep functions and methods small: A function or method should do one thing, and do it well. Keeping them small makes it easier to understand, test and maintain them.

- Use meaningful names: Names of variables, functions, and types should be meaningful and reflect their purpose. Avoid using abbreviations or acronyms that are not widely understood.

- Write clear and concise comments: Comments should be used sparingly and only when necessary. They should explain why something is done, not what is done, as the code should be self-explanatory.

- Write simple and clear code: Avoid writing complex code that is difficult to understand. Simplify the code by breaking it into smaller functions or by using helper functions.

- Use error handling: Proper error handling is essential to write reliable software. Always check for errors and handle them appropriately.

- Keep code formatting consistent: Consistent formatting makes the code easier to read and understand. Use an automated formatting tool like gofmt to ensure that the code adheres to a consistent style.

- Write meaningful and comprehensive tests: Tests should cover all important scenarios in your code. They should be written in a way that is easy to understand and maintain, and should provide clear feedback on what went wrong if they fail.


## The **SOLID** principals   


**Single Responsibility Principle (SRP)**: Each function or module should have only one responsibility, and should do it well. In Go, this means breaking down larger functions into smaller, more focused ones that do one thing.

**Open/Closed Principle (OCP)**: Software entities should be open for extension but closed for modification. In Go, this can be achieved through the use of interfaces, which define a contract for behavior, and can be implemented by different types.

**Liskov Substitution Principle (LSP)**: Subtypes should be substitutable for their base types. In Go, this means that any type that implements an interface should be able to be used in place of that interface without causing any issues.

**Interface Segregation Principle (ISP)**: Clients should not be forced to depend on interfaces they do not use. In Go, this means that interfaces should be small and focused, with only the necessary methods.

**Dependency Inversion Principle (DIP)**: High-level modules should not depend on low-level modules; both should depend on abstractions. In Go, this means defining interfaces for dependencies, rather than depending on concrete types directly.
