# IP-Address-Counter
Golang-based solution for counting uniques IP address records in a large file

<b>1. Introduction</b>

The purpose of the code in this repo is to solve the following problem:

There is a simple text file containing IPv4 addresses, each line of the file represents an IPv4 address. The task is to build a program in Go language that calculates how many unique IP addresses are in this file. The task includes a sample large file (~114 GB) that needs to be properly handled by the program.

The solution was tested on MacBook Pro 2021 with 10 processor cores.

<b>2. Discussion</b>

Starting from the beginning, I have used the suggested large file ip_addresses as well as another small file with 27 lines and 10 unique addresses as testing subjects - small file was used to test correctness, and large file - for performance.

Perhaps, the first reaction when you see such a task, is to check if there is a simple and straightforward, or "naive" solution, if for nothing else but for benchmarking purposes.
The solution that suggests itself is to:

1. Create IP address registry in a form of map[IP_Address]bool
2. Scan through the file line by line and for each line add it to the map using obtained Ip address string as key, and "true" as value.
3. Return the length of the map.

When implemented, this method works well on the small file, but fails miserably when tested on the large one - the execution completely stalls at around 1.3 billion read lines in the file. This was completely expected, since every time we attempting to write to our map-based registry, the program needs to check if the item is already there, and that process grows progressively with increasing size of the map, eventually choking the processing resources.

To overcome this, we can replace the map-based registry with an array that will hold a boolean flag for each possible IP address.

Essentially, an IPv4 address is a group of four numbers from 0 to 256. Therefore, we can create an IP registry in the form of the following four-dimensional array:

ipRegistry := [256][256][256][256]bool{}

Presence of each possible IPv4 address can be represented in this registry by the boolean flag that is addressed accordingly to the four IPv4 numbers.

The problem with this representation is that the bool value means only two states - true or false - but occupies the whole byte, or eight bit, in memory. The total size of the array is therefore will be 256*256*256*256 B = 4.3 GB.

To reduce this size, we could replace the final [256]bool dimension of the registry array with a single int256 value, with each binary bit of it serving us as a flag in lieu of bool.

Unfortunately, Golang does not have a built-in implementation for 256-bit integers, so the next best thing would be to have an array of four 64-bit integers in its place. In this array, each 64-bit integer will be responsible for four consecutive 64-bit groups that make a whole 256 number. The only tricky thing there would be to properly write to this registry, but that could be implemented with bitwise operations.

With this approach, we arrive to to following final version of the IP registry array:

ipRegistry := [256][256][256][4]int64{}

The size of this array is  256*256*256*4*8 = 537 MB, which is much more palatable than the previous version.

The next consideration in optimizing the algorithm would be the time factor. So far we have the following sequence:

1. Initializing the IP registry
2. Scan the file line by line, processing it, recording changes in the IP registry
3. Once the scan is complete, going through the registry and counting the total number of 1-bits in all the int64 values.
4. Reporting the total 1-bit count as the number of unique IP addresses

Once implemented, this algorithm works fine and takes about 16 minutes to process the sample large file. To speed up the process, we may employ concurrent processing in the following way.

Instead of line-by line processing, we could do so in batches. This is how the algorithm could look like:

1. Initializing the IP registry
2. Scan the file for lines, appending each line to the slice of strings.
3. Once the length of the slice achieves certain pre-determined batch size, the slice is passed to batch processor
4. Batch processor receives the batch, splits it to a number of sub-batches, and processes these sub-batches concurrently using Goroutines.
5. Entire batch processing can also be executed concurrently with the scanner continuing building the next batch
6. The process is repeated from step 2 until entire file has been scanned.

<b>3. Implementation</b>

The solution to the problem is based upon the approach discussed above. It is presented in this repo as "ip_address_counter.go" file and can be run in the following way:
go run ip_address_counter.go fileName
where fileName is the name of file that needs to be examined. Also, the file needs to be in the same directory as the program itself.
When tested on the provided sample ip_addresses file, the total execution time was under 6 minutes.
