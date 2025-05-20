import std.stdio;

void main()
{
    for (int i = 0; i < 50; i++)
    if (i == 17 || i == 20 || i == 49)
    {
        write(" NONE ");
    }
    else
    {
        write(i, " ");
    }
    
    // if setningarnar virðast ekki virka hér fyrir neðan en þær
    // virka í fyrri lúppunni
    
    writeln("\n\n");
    
    for (int j = 0; j < 50; j++)
    if (j == 1369)
    {
        write(" NONE ");
    }
    else
    {
        write(j*j, " ");
    }
}