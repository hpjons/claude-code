#include <iostream>
#include <array>
#include <iomanip>

int main() {
    std::array<double, 4> arr = {2.3,4.3,5.4,6.7};
    std::array<double, 6> arr2 {1,2,3,4,5,6};
    std::cout << std::fixed << std::setprecision(2);

    for (double num : arr){
        std::cout << num << " ";
    }

    for (double num2 : arr2){
	    std::cout << num2 << " ";
    }

    std::cout << std::endl;
    return 0;
}
