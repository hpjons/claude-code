program fargmoldun;
var
  i,j : integer;
begin
  j := 1;

  while j <= 10 do
  begin
    i := 0;
      while i <= 10 do
      begin
        write(i*j, '-');
        i := i + 1;
      end;
    writeln;
    j := j + 1;
  end;
end.
