
# Promotions In Go

**Instruction to run**\
Provided file should be named "promotions.csv" and be located in the same diretory as 'app.go'.
PostgreSQL should be installed, to which the system connects with "postgres" username and "123" password. The db name is "postgres".

**DB Indexing**\
Decided to go with the BTree indexing, even though BRIN also looked promising.
But BRIN works best with sorted tables, and I am not sure if sorting a billion entries is worth it?

**Reading the CSV file**\
Tried multiple approaches for reading the csv file
1. Reading the whole file, then parsing it into list of objects, the issue with this solution was that if the file is too large the process memory will be filled up.
2. Reading the file line by line (which worked the best for the provided file containing 200.000 entries) and then parsing & appending each line to the list of objects.
3. Reading a file with consumer/producer model, but this solution didn't have any visible advantage over the second solution. Maybe for bigger files the advantage will become more obvious so leaving it in draft.

**Insertion**\
Used Unnest method to Insert data into DB, though considered CopyIn also:

1. Unnest: which allows to insert more than 65k+ queries at a time
2. CopyIn: which does the copying automatically, but is a bit riskier, especially considering the .csv files do not have
headers

Both methods execute in about the same time.

**Available Methods**\
Created the following method according to the task.
|  |  |
|--|--|
| GET    /promotions/{id}  | _gets promotion by id_ |

Added extra methods for internal testing.

| Call                                              | _Description_             |
|---------------------------------------------------|---------------------------|
| GET    /promotions                                | _gets 50 promotions_      |
| POST   /promotions                                | _create new promotion_    |
| PUT    /promotions/{id}                           | _updates promotion_       |
| DELETE /promotions/{id}                           | _deletes promotion by id_ |


**Load Testing**

The testing was done using [Vegeta](https://github.com/tsenart/vegeta).\
Calling the server for 5s with a rate of 600. The result are the following

    Requests      [total, rate, throughput]         5000, 1000.21, 1000.19
    Duration      [total, attack, wait]             4.999s, 4.999s, 93µs
    Latencies     [min, mean, 50, 90, 95, 99, max]  48.4µs, 97.039µs, 69.317µs, 86.461µs, 93.972µs, 512.214µs, 47.911ms
    Bytes In      [total, mean]                     568906, 113.78
    Bytes Out     [total, mean]                     0, 0.00
    Success       [ratio]                           100.00%
    Status Codes  [code:count]                      200:5000
    Error Set:

Calling the server for 10s with a rate of 1500. The result are the following

    Requests      [total, rate, throughput]         15000, 1500.01, 1500.01
    Duration      [total, attack, wait]             10s, 10s, 0s
    Latencies     [min, mean, 50, 90, 95, 99, max]  0s, 1.019ms, 77.052µs, 95.236µs, 109.32µs, 1.934ms, 342.296ms
    Bytes In      [total, mean]                     1706621, 113.77
    Bytes Out     [total, mean]                     0, 0.00
    Success       [ratio]                           100.00%
    Status Codes  [code:count]                      200:15000
    Error Set:

Calling the server for 10s with a rate of 3000. The result are the following

    Requests      [total, rate, throughput]         30000, 3000.17, 1800.00
    Duration      [total, attack, wait]             16.095s, 9.999s, 6.096s
    Latencies     [min, mean, 50, 90, 95, 99, max]  80.9µs, 262.986ms, 80.962µs, 1.1ms, 2.33s, 5.74s, 7.427s
    Bytes In      [total, mean]                     3418649, 113.95
    Bytes Out     [total, mean]                     0, 0.00
    Success       [ratio]                           96.57%
    Status Codes  [code:count]                      200:28971  500:1029
    Error Set:
    500 Internal Server Error{"error":"dial tcp [::1]:5432: connectex: No connection could be made because the target machine actively refused it."}
