import std.stdio;

void main()
{


    int[string] books = [
        1: "\tCosmos\n\tThe Demon Haunted World\n\tPale Blue Dot\n",
        2: "\tRíkið\n",
        3: "\tPride and Prejudice\n",
        4: "\tThe C Programming Language",
    ];


foreach (author, bookList; books) {
    writeln(author, bookList);
}

}
