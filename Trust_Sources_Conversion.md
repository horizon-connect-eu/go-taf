
### Formulas for TS conversion

$\omega_{DTI} = (b_{DTI}, d_{DTI}, u_{DTI})$ - **initial trust opinion** based on the Design Time Information
$\omega_{RTI} = (b_{RTI}, d_{RTI}, u_{RTI})$ - **run-time trust opinion** calculated during run-time
$w_x$ - the **weight** associated with trust source $TS_x$


Calculating the **delta** which will be used for adjusting the trust opinion:
$\Delta_x = w_x * u_{DTI}$ 

##### Running the conversion the 1st time

Calculating the **belief**, **disbelief**, and **uncertainty** after processing evidence from a trust source $TS_x$ for the first time, where $x \in [1, 2, 3, ...n)$, $n$ is the total number of trust sources (e.g. 3 in this case), and $TS_x$ is a boolean: 

$b_x = b_{x-1} + TS_x*\Delta_x$
$d_x = d_{x-1} + (1-TS_x)*\Delta_x$
$u_x = u_{x-1} - \Delta_x$

$\omega_{RTI} = (b_n, d_n, u_n)$

Note that in this case:
$b_0 = b_{DTI}$, $d_0 = d_{DTI}$, and $u_0 = u_{DTI}$ 

Moreover, note that in this case, all of the trust sources are taken into account.
##### Running the conversion any later time

When a run-time trust opinion has been calculated and we need to re-assess it when there is a change caused by new evidence from the trust sources we have already gotten evidence from, we use the following formulas:

$b_x = b_{x-1} + (TS_{x, new} - TS_{x, old})*\Delta_x$
$d_x = d_{x-1} - (TS_{x, new} - TS_{x, old})*\Delta_x$
$u_x = u_{x-1}$

Note that in this case:
$b_0 = b_{RTI}$, $d_0 = d_{RTI}$, and $u_0 = u_{RTI}$ 

Note $TS_{x, new} - TS_{x, old}$ options: 
0 - 1 = -1 (negative change)
0 - 0 = 0 (no change)
1 - 1 = 0 (no change)
1 - 0 = 1 (positive change)

In this case, the belief or disbelief have to change depending on whether the change in evidence is positive or negative. If the change is negative, meaning we go from positive evidence to negative evidence, then the belief should drop and the disbelief should increase. If the change is positive, meaning we go from negative evidence to positive evidence, the the belief should increase, and the disbelief should increase. The uncertainty should stay the same unless there is evidence from a trust source we haven't processed yet, in which case we use the formulas from "Running the conversion for the 1st time".

### Example of conversion

We are focusing on the DENSO ECU migration use case and we use the following trust model in this case:
![[TAF_Brussels_TM2.png]]]

We will be assessing the trustworthiness of an ECU as assessed by the TAF, e.g. $\omega^{TAF}_{ECU_1}$.

The initial trust opinion in this trust relationship: $\omega_{DTI} = (0.25, 0.18, 0.57)$

Types of Trust Sources that we will use:
- **static** = do not change during runtime
- **dynamic** = ***do*** change during runtime

Trust Sources to be used:
1. $TS_1$: Secure Boot (**static**)
	- weight $w_1 = 0.2$ 
2. $TS_2$: Dynamic Control Flow Integrity (**dynamic**)
	- weight $w_2 = 0.4$
1. $TS_3$: Intrusion Detection System (**dynamic**)
	- weight $w_3 = 0.4$

Therefore, for these three trust sources, the deltas are:
$\Delta_1 = 0.2 * 0.57 = 0.114$
$\Delta_2 = 0.4 * 0.57 = 0.228$
$\Delta_3 = 0.4 * 0.57 = 0.228$

**Case 1**: Initially, all 3 trust sources provide **positive** evidence.

1. $TS_1 = 1$     $\Rightarrow$     $b_1 = 0.25 + 0.114 = 0.364$  $|$ $d_1 = 0.18$ $|$ $u_1 = 0.57 - 0.114 = 0.456$
2. $TS_2 = 1$     $\Rightarrow$     $b_2 = 0.364 + 0.228 = 0.592$ $|$ $d_2 = 0.18$ $|$ $u_2 = 0.456 - 0.228 = 0.228$ 
3. $TS_3 = 1$     $\Rightarrow$     $b_3 = 0.592 + 0.228 = 0.820$ $|$ $d_3 = 0.18$ $|$ $u_3 = 0.228 - 0.228 = 0$

Therefore, the run-time Trust Opinion after the first conversion is $\omega_{RTI} = (0.820,0.18,0)$. 

Therefore, $$\omega^{TAF}_{ECU_1}=(0.820,0.18,0)$$

**Case 2**: At some point, a single trust source will provide **negative** evidence.

1. $TS_1 = 1$     $\Rightarrow$     $b_1 = 0.820+(1 - 1)\Delta_1 = 0.820$  $|$ $d_1 = 0.18 - (1 - 1) * \Delta_1 = 0.18$ $|$ $u_1 = 0$
2. $TS_2 = 1$     $\Rightarrow$     $b_2 = 0.820 + (0 - 1)*0.228 = 0.592$ $|$ $d_2 = 0.18 - (0 - 1) * 0.228 = 0.408$ $|$ $u_2 = 0$ 
3. $TS_3 = 1$ 	  $\Rightarrow$		$b_3 = 0.592 + (1-1) * \Delta_3 = 0.592$ $|$ $d_3 = 0.408 - (1-1) * \Delta_3 = 0.408$ $|$ $u_3 = 0$

Therefore, the run-time Trust Opinion after the second conversion is $\omega_{RTI} = (0.592,0.402,0)$. 

Therefore, $\omega^{TAF}_{ECU_1} = (0.592,0.402,0)$.
