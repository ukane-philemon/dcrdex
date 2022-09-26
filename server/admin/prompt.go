// This code is available on the terms of the project LICENSE.md file,
// also available online at https://blueoakcouncil.org/license/1.0.0.

package admin

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"

	"decred.org/dcrdex/dex/encode"
	"golang.org/x/term"
)

type passwordReadResponse struct {
	password []byte
	err      error
}

// PasswordPrompt prompts the user to enter a password. Password must not be
// empty.
func PasswordPrompt(ctx context.Context, prompt string) ([]byte, error) {
	// Get the initial state of the terminal.
	initialTermState, err := term.GetState(int(os.Stdin.Fd()))
	if err != nil {
		return nil, err
	}

	passwordReadChan := make(chan passwordReadResponse, 1)
	passwordResult := func(pass []byte, err error) {
		passwordReadChan <- passwordReadResponse{
			password: pass,
			err:      err,
		}
	}

	go func() {
		// Use a retry loop to ensure user password returned from the terminal
		// is not a zero-length byte because it is not an error if len(password)
		// == 0, user did not enter a password.
		var pass []byte
		for len(pass) == 0 {
			fmt.Print(prompt)
			pass, err = term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println()
			if err != nil {
				passwordResult(pass, err)
				return
			}
		}
		passwordResult(pass, nil)
	}()

	select {
	case <-ctx.Done():
		_ = term.Restore(int(os.Stdin.Fd()), initialTermState)
		return nil, ctx.Err()

	case res := <-passwordReadChan:
		if res.err != nil {
			return nil, res.err
		}
		return res.password, nil
	}
}

// PasswordHashPrompt prompts the user to enter a password and returns its
// SHA256 hash. Password must not be an empty string.
func PasswordHashPrompt(ctx context.Context, prompt string) ([sha256.Size]byte, error) {
	var authSHA [sha256.Size]byte
	passBytes, err := PasswordPrompt(ctx, prompt)
	if err != nil {
		return authSHA, err
	}
	authSHA = sha256.Sum256(passBytes)
	// Zero password bytes.
	encode.ClearBytes(passBytes)
	return authSHA, nil
}
