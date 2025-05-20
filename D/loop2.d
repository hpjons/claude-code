import std.stdio;

void main()
{
    inner(1);
}

int inner(int x)
{
    int counter = x;
    
    for(int i = 1; i < 5; i++)
    {
        write(counter);
        counter = counter + i;
    }
    return counter;
}