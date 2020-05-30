package main

import (
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/k0kubun/go-ansi"
	"github.com/mitchellh/colorstring"
)

// ---"basic" types ---
type basicOpSolver func(n1, n2 int) (solution int)
type basicOp struct {
	solver basicOpSolver
	sign   string
}

func (e basicOp) String() string {
	return e.sign
}

// basic Expression. 2 numbers, 1 operator, 1 solution
type basicExpr struct {
	num1, num2 int
	op         basicOp
}

func (e basicExpr) String() string {
	return fmt.Sprintf("%d %s %d", e.num1, e.op.sign, e.num2)
}

func (e basicExpr) Answer() int {
	return e.op.solver(e.num1, e.num2)
}

// --- "basic" operations ---

func add(n1, n2 int) int {
	return n1 + n2
}

func multiply(n1, n2 int) int {
	return n1 * n2
}

var addition = basicOp{solver: add, sign: "+"}
var multiplication = basicOp{solver: multiply, sign: "x"}

// --- logic ---

func randSelect(set []interface{}) interface{} {
	return set[rand.Intn(len(set))]
}

func randRange(low, high int) int {
	return rand.Intn(high-low) + low
}

func generateExpression(genNum1, genNum2 func() int, ops []basicOp) basicExpr {
	return basicExpr{
		num1: genNum1(),
		num2: genNum2(),
		op:   ops[randRange(0, len(ops))],
	}
}

func testBasicExpr(expr basicExpr) (correct bool, msAnswerTime int, userAnswer int) {
	var userAns int
	startTime := time.Now()
	for {
		fmt.Print("     ", expr, " = ")
		n, err := fmt.Scanln(&userAns)
		if err != nil {
			var t string
			fmt.Scanln(&t)
			terminalLineUp()
			terminalClear()
			continue
		}
		if n != 1 {
			panic("couldn't read user answer")
		}
		break
	}
	return userAns == expr.Answer(), int(time.Since(startTime).Milliseconds()), userAns
}

func trainBasicExpr(exgen func() basicExpr, count int, slowThresh int) {
	if count <= 0 {
		return
	}
	type response struct {
		q              basicExpr
		isCorrect      bool
		answerTimeMili int
		userAnswer     int
		feedback       string
	}
	responses := make([]response, count)
	for i := 0; i < count; i++ {
		ex := exgen()
		isCorrect, msAswerTime, userAnswer := testBasicExpr(ex)
		// delete quetion (print it with later with answer)
		terminalLineUp()
		terminalClear()
		var status, comment, color string
		if isCorrect {
			status = "GOOD"
			comment = ""
			color = "[green]"
		} else {
			status = "BAD "
			comment = fmt.Sprintf("(you said %d)", userAnswer)
			color = "[red]"
		}

		s := fmt.Sprintf("%s%s %v = %d %s", color, status, ex, ex.Answer(), comment)
		colorstring.Fprintln(terminalANSI, s)

		responses[i] = response{ex, isCorrect, msAswerTime, userAnswer, s}
	}
	fmt.Println("Done")

	// we know the max size :)
	var wrong []response
	var slow []response
	averageTime := float64(responses[0].answerTimeMili)

	for _, resp := range responses {
		if !resp.isCorrect {
			wrong = append(wrong, resp)
		}
		if resp.answerTimeMili > slowThresh {
			slow = append(slow, resp)
		}
		averageTime = (averageTime + float64(resp.answerTimeMili)) / 2.0
	}

	correctCnt := count - len(wrong)
	colorstring.Fprintln(terminalANSI, "[cyan]Stats:")
	if correctCnt == count {
		colorstring.Fprintln(terminalANSI, "[cyan]PERFECT SCORE!")
		colorstring.Fprintln(terminalANSI, "[cyan]PERFECT SCORE!")
		colorstring.Fprintln(terminalANSI, "[cyan]PERFECT SCORE!")
		colorstring.Fprintln(terminalANSI, "[cyan]PERFECT SCORE!")
		colorstring.Fprintln(terminalANSI, "[cyan]PERFECT SCORE!")
	}
	colorstring.Fprintln(terminalANSI, fmt.Sprintf(" [green]Correct: %d/%d (%0.2f%%)", correctCnt, count, float64(correctCnt)/float64(count)*100))
	colorstring.Fprintln(terminalANSI, fmt.Sprintf(" [yellow]Slow: %d/%d (%0.2f%%)", len(slow), count, float64(len(slow))/float64(count)*100))
	colorstring.Fprintln(terminalANSI, fmt.Sprintf(" [red]Wrong: %d/%d (%0.2f%%)", len(wrong), count, float64(len(wrong))/float64(count)*100))
	if len(slow) > 0 {
		colorstring.Fprintln(terminalANSI, "")
		colorstring.Fprintln(terminalANSI, "Exercises with [yellow]slow[reset] answers:")
		for _, s := range slow {
			colorstring.Fprintln(terminalANSI, s.feedback)
		}

	}
	if len(wrong) > 0 {
		colorstring.Fprintln(terminalANSI, "")
		colorstring.Fprintln(terminalANSI, "Exercises with [red]mistakes[reset]:")
		for _, w := range wrong {
			colorstring.Fprintln(terminalANSI, w.feedback)
		}
	}
}

func terminalLineUp() {
	ansi.CursorUp(1)
}
func terminalClear() {
	fmt.Print("\r")
}

func printMultiplicationTable(min, max, colorFrom int) {
	clrprnt := func(s string) {
		colorstring.Fprint(terminalANSI, s)
	}
	// make border
	prefix := "[cyan]  *"
	clrprnt(prefix)
	for j := min; j < max; j++ {
		clrprnt(fmt.Sprintf("[magenta]%5d", j))
	}
	fmt.Println()

	var color string
	for i := min; i < max; i++ {
		clrprnt(fmt.Sprintf("[magenta] %2d", i))
		for j := min; j < max; j++ {
			if i > colorFrom || j > colorFrom {
				color = "[green]"
			} else {
				color = ""
			}
			clrprnt(fmt.Sprintf("%s%5d", color, j*i))
		}
		fmt.Println()
	}
}

var terminalANSI io.Writer

func main() {
	// initing
	rand.Seed(time.Now().UTC().UnixNano())
	supportedOperations := [...]basicOp{addition, multiplication}
	selectedOperations := supportedOperations[1:]
	slowThresh := 2 * 1000 // slow answer if > 3 secodns
	exCount := 100
	// MAX NUMS are not including
	n1MinNum := 3
	n1MaxNum := 12
	n2MinNum := 6
	n2MaxNum := 12
	fmt.Println("Welcome")
	fmt.Println("Current supported operations:", supportedOperations)
	fmt.Println("Currently selected operations:", selectedOperations)
	fmt.Println("Slow Threshold:", slowThresh, "ms")
	fmt.Println("num1 range:", n1MinNum, "-", n1MaxNum)
	fmt.Println("num2 range:", n2MinNum, "-", n2MaxNum)

	terminalANSI = ansi.NewAnsiStdout()
	if false {
		printMultiplicationTable(4, 15, 10)
		return
	}
	genNum1 := func() int {
		return randRange(n1MinNum, n1MaxNum)
	}
	genNum2 := func() int {
		return randRange(n2MinNum, n2MaxNum)
	}

	exerciseGenerator := func() basicExpr {
		return generateExpression(genNum1, genNum2, selectedOperations)
	}
	trainBasicExpr(exerciseGenerator, exCount, slowThresh)
}
