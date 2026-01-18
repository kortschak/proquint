// https://arxiv.org/html/0901.4016
package main

import (
	"bufio"
	"crypto/rand"
	"flag"
	"fmt"
	"math/big"
	"os"
	"strings"
	"unicode"

	"upspin.io/key/proquint"
)

func main() {
	rndBits := flag.Uint("r", 0, "generate random proquint phrase with given bit length if non-zero")
	flag.Parse()

	if *rndBits != 0 {
		max := big.NewInt(1)
		max.Lsh(max, 16*(*rndBits/16+1))
		i, err := rand.Int(rand.Reader, max)
		if err != nil {
			os.Exit(1)
		}
		fmt.Println(encode(i, 0))
	}

	if flag.NArg() != 1 {
		return
	}

	arg, err := input(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if isNumber(arg) {
		n, b := leadingZeros(arg)
		var i big.Int
		err := i.UnmarshalText(b)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}

		fmt.Println(encode(&i, n))
		return
	}

	n := decode(arg)
	if n == "" {
		fmt.Fprintln(os.Stderr, "invalid proquint:", arg)
		os.Exit(2)
	}
	fmt.Println(n)
}

func input(s string) (string, error) {
	if s != "-" {
		return s, nil
	}
	sc := bufio.NewScanner(os.Stdin)
	if sc.Scan() {
		return sc.Text(), nil
	}
	return s, sc.Err()
}

func leadingZeros(s string) (int, []byte) {
	s, isHex := strings.CutPrefix(s, "0x")
	var prefix string
	if isHex {
		prefix = "0x"
	}
	for i, r := range s {
		if r != '0' {
			return i, []byte(prefix + s[i:])
		}
	}
	return len(s), nil
}

func isNumber(s string) bool {
	s, isHex := strings.CutPrefix(s, "0x")
	for _, r := range s {
		r = unicode.ToLower(r)
		if (r < '0' || '9' < r) && (!isHex || (r < 'a' || 'f' < r)) {
			return false
		}
	}
	return true
}

func proquintWords(i int) int {
	return (i-1)/16 + 1
}

func encode(i *big.Int, leadingZeros int) string {
	words := make([]string, proquintWords(i.BitLen()))
	mask := big.NewInt(0xffff)
	var t big.Int
	for w := range words {
		words[w] = string(proquint.Encode(uint16(t.And(i, mask).Uint64())))
		i.Rsh(i, 16)
	}
	for i, j := 0, len(words)-1; i < j; i, j = i+1, j-1 {
		words[i], words[j] = words[j], words[i]
	}
	prefix := make([]string, leadingZeros)
	for i := range prefix {
		prefix[i] = string(proquint.Encode(0))
	}
	return strings.Join(append(prefix, words...), "-")
}

func decode(s string) string {
	var numbers []uint16
	for _, w := range strings.Split(s, "-") {
		if len(w) != 5 {
			return ""
		}
		if !isValidProquint(w) {
			return ""
		}
		numbers = append(numbers, proquint.Decode([]byte(w)))
	}
	zeros := 0
	for i, n := range numbers {
		if n != 0 {
			if i != 0 {
				fmt.Print(strings.Repeat("0", i))
				zeros = i
			}
			break
		}
	}
	var i big.Int
	var t big.Int
	for w := zeros; w < len(numbers); w++ {
		i.Lsh(&i, 16)
		i.Or(&i, t.SetUint64(uint64(numbers[w])))
	}
	return i.Text(10)
}

func isValidProquint(s string) bool {
	return len(s) == 5 && con[s[0]] && vo[s[1]] && con[s[2]] && vo[s[3]] && con[s[4]]
}

var (
	con = [0x100]bool{
		'b': true, 'd': true, 'f': true, 'g': true,
		'h': true, 'j': true, 'k': true, 'l': true,
		'm': true, 'n': true, 'p': true, 'r': true,
		's': true, 't': true, 'v': true, 'z': true,
	}
	vo = [0x100]bool{
		'a': true, 'i': true, 'o': true, 'u': true,
	}
)
