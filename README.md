# Promotions In Go

Went with PostgreSQL database, 


**DB Indexing**\
Decided to go with the BTree indexing, even though BRIN also looked promising.
But BRIN works best with sorted tables, and I am not sure if sorting a billion entries is worth it?

**Insertion**\
Used two different types to Insert data into DB:
1. unnest: which allows to insert more than 65k+ queries at a time
2. CopyIn: which does the copying automatically, but is a bit riskier, especially considering the .csv files do not have headers

Both methods execute in about the same time.