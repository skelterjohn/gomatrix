// Copyright 2009 The GoMatrix Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package matrix

import (
	"fmt";
	"rand";
)

/*
A sparse matrix based on go's map datastructure.
*/
type SparseMatrix struct {
	matrix;
	elements	map[int]float64;
	// offset to start of matrix s.t. idx = i*cols + j + offset
	// offset = starting row * step + starting col
	offset	int;
	// analogous to dense step
	step	int;
}

func (A *SparseMatrix) Get(i int, j int) float64 {
	x, ok := A.elements[i*A.step+j+A.offset];
	if !ok {
		return 0
	}
	return x;
}

/*
Looks up an element given its element index.
*/
func (A *SparseMatrix) GetIndex(index int) float64 {
	x, ok := A.elements[index];
	if !ok {
		return 0
	}
	return x;
}

/*
Turn an element index into a row number.
*/
func (A *SparseMatrix) GetRowIndex(index int) int {
	return (index - A.offset) / A.cols
}

/*
Turn an element index into a column number.
*/
func (A *SparseMatrix) GetColIndex(index int) int {
	return (index - A.offset) % A.cols
}

/*
Turn an element index into a row and column number.
*/
func (A *SparseMatrix) GetRowColIndex(index int) (i int, j int) {
	i = (index - A.offset) / A.step;
	j = (index - A.offset) % A.step;
	return;
}

func (A *SparseMatrix) Set(i int, j int, v float64) {
	// v == 0 results in removal of key from underlying map
	A.elements[i*A.step+j+A.offset] = v, v != 0
}

/*
Sets an element given its index.
*/
func (A *SparseMatrix) SetIndex(index int, v float64) {
	// v == 0 results in removal of key from underlying map
	A.elements[index] = v, v != 0
}

/*
A channel that will carry the indices of non-zero elements.
*/
func (A *SparseMatrix) Indices() (out chan int) {
	//maybe thread the populating?
	for index := range A.elements {
		out <- index
	}
	return;
}

/*
Get a matrix representing a subportion of A. Changes to the new matrix will be
reflected in A.
*/
func (A *SparseMatrix) GetMatrix(i int, j int, rows int, cols int) *SparseMatrix {
	B := new(SparseMatrix);
	B.rows = rows;
	B.cols = cols;
	B.offset = (i+A.offset/A.step)*A.step + (j + A.offset%A.step);
	B.step = A.step;
	B.elements = A.elements;
	return B;
}

/*
Gets a reference to a column vector.
*/
func (A *SparseMatrix) GetColVector(j int) *SparseMatrix {
	return A.GetMatrix(0, j, A.rows, j+1)
}

/*
Gets a reference to a row vector.
*/
func (A *SparseMatrix) GetRowVector(i int) *SparseMatrix {
	return A.GetMatrix(i, 0, i+1, A.cols)
}

/*
Creates a new matrix [A B].
*/
func (A *SparseMatrix) Augment(B *SparseMatrix) (*SparseMatrix, *error) {
	if A.rows != B.rows {
		return nil, NewError(ErrorDimensionMismatch)
	}
	C := ZerosSparse(A.rows, A.cols+B.cols);

	for index, value := range A.elements {
		i, j := A.GetRowColIndex(index);
		C.Set(i, j, value);
	}

	for index, value := range B.elements {
		i, j := B.GetRowColIndex(index);
		C.Set(i, j+A.cols, value);
	}

	return C, nil;
}

/*
Creates a new matrix [A;B], where A is above B.
*/
func (A *SparseMatrix) Stack(B *SparseMatrix) (*SparseMatrix, *error) {
	if A.cols != B.cols {
		return nil, NewError(ErrorDimensionMismatch)
	}
	C := ZerosSparse(A.rows+B.rows, A.cols);

	for index, value := range A.elements {
		i, j := A.GetRowColIndex(index);
		C.Set(i, j, value);
	}

	for index, value := range B.elements {
		i, j := B.GetRowColIndex(index);
		C.Set(i+A.rows, j, value);
	}

	return C, nil;
}

/*
Returns a copy with all zeros above the diagonal.
*/
func (A *SparseMatrix) L() *SparseMatrix {
	B := ZerosSparse(A.rows, A.cols);
	for index, value := range A.elements {
		i, j := A.GetRowColIndex(index);
		if i >= j {
			B.Set(i, j, value)
		}
	}
	return B;
}

/*
Returns a copy with all zeros below the diagonal.
*/
func (A *SparseMatrix) U() *SparseMatrix {
	B := ZerosSparse(A.rows, A.cols);
	for index, value := range A.elements {
		i, j := A.GetRowColIndex(index);
		if i <= j {
			B.Set(i, j, value)
		}
	}
	return B;
}

func (A *SparseMatrix) Copy() *SparseMatrix {
	B := ZerosSparse(A.rows, A.cols);
	for index, value := range A.elements {
		B.elements[index] = value
	}
	return B;
}

func ZerosSparse(rows int, cols int) *SparseMatrix {
	A := new(SparseMatrix);
	A.rows = rows;
	A.cols = cols;
	A.offset = 0;
	A.step = cols;
	A.elements = map[int]float64{};
	return A;
}

/*
Creates a matrix and puts a standard normal in n random elements, with replacement.
*/
func NormalsSparse(rows int, cols int, n int) *SparseMatrix {
	A := ZerosSparse(rows, cols);
	for k := 0; k < n; k++ {
		i := rand.Intn(rows);
		j := rand.Intn(cols);
		A.Set(i, j, rand.NormFloat64());
	}
	return A;
}

/*
Create a sparse matrix using the provided map as its backing.
*/
func MakeSparseMatrix(elements map[int]float64, rows int, cols int) *SparseMatrix {
	A := ZerosSparse(rows, cols);
	A.elements = elements;
	return A;
}

/*
Convert this sparse matrix into a dense matrix.
*/
func (A *SparseMatrix) DenseMatrix() *DenseMatrix {
	B := Zeros(A.rows, A.cols);
	for index, value := range A.elements {
		i, j := A.GetRowColIndex(index);
		B.Set(i, j, value);
	}
	return B;
}

func (A *SparseMatrix) String() string {

	s := "{";
	for i := 0; i < A.Rows(); i++ {
		for j := 0; j < A.Cols(); j++ {
			s += fmt.Sprintf("%f", A.Get(i, j));
			if i != A.Rows()-1 || j != A.Cols()-1 {
				s += ","
			}
			if j != A.cols-1 {
				s += " "
			}
		}
		if i != A.Rows()-1 {
			s += "\n"
		}
	}
	s += "}";
	return s;
}
