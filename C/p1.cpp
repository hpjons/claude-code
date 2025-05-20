#include <iostream>
using namespace std;

int main()
{
    float M_sodi;
    float M_syra;
    float d_syra;
    float vol_syra;
    float m_14prosent_syra;
    float m_sodi;
    float n_syra;
    float n_sodi;

	M_sodi = 84.00;
    M_syra = 60.02;
    d_syra = 1.049;
    vol_syra = 10;
    m_14prosent_syra = vol_syra * d_syra * 0.14;
    m_sodi = 4;

    n_syra = m_14prosent_syra/M_syra;
    n_sodi = m_sodi/M_sodi;

    cout << "massi 14% sýru : " << m_14prosent_syra << endl;
    cout << "mólmagn sýru : " << n_syra << endl;
    cout << "mólmagn sóda : " << n_sodi << endl;

    return 0;
    
}
