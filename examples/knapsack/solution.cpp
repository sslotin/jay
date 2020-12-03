#include <bits/stdc++.h>
using namespace std;

int main() {
    int n;
    cin >> n;

    vector<int> a(n);
    
    for (int &x : a) {
        cin >> x;
    }
    
    int ans = 0;
    for (int mask = 1; mask < (1<<n); mask++) {
        int s = 0;
        for (int i = 0; i < n; i++) {
            if (mask>>i&1) {
                s += a[i];
            }
        }
        if (s % n == 0) {
            ans = mask;
        }
    }

    for (int i = 0; i < n; i++) {
        if (ans>>i&1) {
            cout << i + 1 << " ";
        }
    }
    
    cout << endl;


    return 0;
}