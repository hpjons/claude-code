import std.stdio;

class Calculator(int a, int b) {
    int add(int a, int b){
        return a + b;
    }
}
    
void main(){
    Calculator calc = new Calculator();

    int result = calc.add(10, 15);
    writeln("Summan er: ", result);
}
