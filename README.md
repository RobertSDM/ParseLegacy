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

The first idea that came into my mind was to use python language.
I was going to deal with tabulated data and the `pandas` library seemed to be just right.

Developing the project in python was straightforward.
The problem really showed up when creating the executables.
The program had to be build specifically for a windows 7 with x86 and x64 as architectures,
and I wanted the build to be generated and published as a release with a Github's Action.

I wasn't being able to create the x86 builds using the `pyinstaller` library.
I keeped getting errors building for x86, and I needed to downgrade the python version to one that still supported windows 7,
then had dependency errors. It was being more complicated then it really needed to be.

The main problem was to build for x86 architecture,
which pyinstaller does, but only if executed in a x86 environment.
The solution was to change the programming language to one where the build compilation would be easier.

Golang was my choice, because it can compile to x86 and other architectures by just changing a environment variable.
It was very easy to integrate in a Github's Action. After deciding the language to use I had to update the code to use GO 1.20 version and adjust some dependencies
and start to call the Windows API.
