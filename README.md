# Parse Legacy

This project is an ETL (Extract Transform and Load), created to translate data represented as a table on a legacy terminal system on a Distribution Center.

## Extraction

Extracting the data was a fun exercise.

```
                                          Qtd.         Qtd.
 Loja   Fabr   Prod                      Pedida       Receb.       Qtd.Corte      Data          Hora   Usuario
 2      5013  51421                           2            1               2      26/11/2025    07:22  29327047
 40     8224  54620                           2                                   26/11/2025    07:22  18310075
 93     7919  51026                           2            1                      26/11/2025    07:22  34298929
 90     6695  52308                           2                                   26/11/2025    07:22  30746860
 75     8582  33500                           4            3                      26/11/2025    07:22  21743481
 74     7536  32850                           2            1               6      26/11/2025    07:22  23490784
 65     8785  57220                           2            1                      26/11/2025    07:22  35378475
 40     7105  26189                           5            3                      26/11/2025    07:22  23258579
 18     7223  44884                           4                           10      26/11/2025    07:22  14395822
 88     7225  56879                           2                                   26/11/2025    07:22  32379928
 44     7262  26114                           3            1                      26/11/2025    07:22  30530836
 89     6575  48318                           1                                   26/11/2025    07:22  27227650
 39     6603  38262                           1                                   26/11/2025    07:22  17905984
 61     8943  31352                           1                                   26/11/2025    07:22  10087555
                   TOTAL=>

-----------------------------------------------------------------------------------------------------------------

```

At first look, the extraction is simple. Take a line and split it by spaces, assign the values to the corresponding columns by order.

But in the exemple above you can see that the certain columns can have null values.

```
 18     7223  44884                           4         [missing]         10      26/11/2025    07:22  14395822
```

Now the solution can be summarized as: get all columns values in order, and identify correctly the ones that are missing. The main problem is the nonexistent separators to identify the column's start and end.

I knew that there had to be a pattern to exhibit this table in a understanding way, and I found it. It was the column alignment by data type!

```
                                      Qtd.          Qtd.
 |Loja   Fabr|   Prod|              Pedida|       Receb.|       Qtd.Corte|     |Data         |Hora  |Usuario
  2      5013   51421                    2             1                2       26/11/2025    07:22  29327047
  40     8224   54620                    2                                      26/11/2025    07:22  18310075

----------------------------------------------------------------------------------------------------------------

```

In the table above, the mark "|" on the side of the column headers indicate the side where the values align. As shown above, the mark can be at the left or right side of the column. In this program the only side that mathers is the right side, because it shows the column's end.

The execution evolves taking all headers alignments positions and extracting the data

```go
columsNamesPositions := parseLegacy.HeadersPositions(columnsNamesLine)
tb := parseLegacy.ParseTable(strTableRows, columsNamesPositions)
```

## Transform

The amount of data the project will extract from the legacy system, can be loaded all in memory, without any issues.

## Load

The data is then saved in disk at a specified path, as a .xlsx

## Why GO

I needed to run this project on a Windows 7 86x, and GO with its capability of creating .exe files from different environments it was running on, using only environment variables, was chosen. I tried with Python, but the .exe creation with pyinstaller was a unecessary burden. I opted by the fast and simple.
