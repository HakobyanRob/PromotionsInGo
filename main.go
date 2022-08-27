package main

func main() {
	a := App{}
	a.Initialize("postgres", "123", "postgres")

	a.Run(":5010")
}
