package main

import (
	"sort"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

var students []Student
var nextID = 1

func findStudentIndex(id int) int {
	for i := range students {
		if students[i].ID == id {
			return i
		}
	}
	return -1
}

// pas searchnya kosong nanti munculin semua cocok. pas searchnya ada nanti bakalan cek nama mahasiswa
func cocokPencarian(s Student, kata string) bool {
	return strings.Contains(strings.ToLower(s.Name), strings.ToLower(kata))
}

func paramID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

// ini fungsi untuk GET /api/v1/students
func listStudents(c *fiber.Ctx) error {
	q := parseListQuery(c)
	hasil := []Student{}

	// nyaring data sesuai query search dan is_active
	for _, s := range students {
		if q.IsActive != nil && s.IsActive != *q.IsActive { continue }
		if q.Search != "" && !cocokPencarian(s, q.Search) { continue }
		hasil = append(hasil, s)
	}

	// ngurutin data sesuai query sort dan order
	sort.SliceStable(hasil, func(i, j int) bool {
		var lebihKecil bool
		switch q.Sort {
		case "nim": lebihKecil = hasil[i].NIM < hasil[j].NIM
		case "name": lebihKecil = hasil[i].Name < hasil[j].Name
		case "grade": lebihKecil = hasil[i].Grade < hasil[j].Grade
		default: lebihKecil = hasil[i].ID < hasil[j].ID
		}
		if q.Order == "desc" { return !lebihKecil }
		return lebihKecil
	})

	// Paginasi
	total := len(hasil)
	totalPages := (total + q.Limit - 1) / q.Limit
	mulai := (q.Page - 1) * q.Limit
	if mulai > total { mulai = total }
	akhir := mulai + q.Limit
	if akhir > total { akhir = total }

	return okList(c, "daftar mahasiswa berhasil diambil", hasil[mulai:akhir], &Meta{
		Page: q.Page, Limit: q.Limit, Total: total, TotalPages: totalPages,
	})
}

// untuk fungsi GET /api/v1/students/:id
func getStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid { return fail(c, fiber.StatusBadRequest, "id harus angka") }

	i := findStudentIndex(id)
	if i == -1 { return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan") } // 404[cite: 1]

	return ok(c, "mahasiswa ditemukan", students[i])
}

// POST /api/v1/students
func createStudent(c *fiber.Ctx) error {
	var req CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus JSON yang valid")
	}

	errs := map[string]string{}
	req.Name = strings.TrimSpace(req.Name)
	req.NIM = strings.TrimSpace(req.NIM)

	if req.Name == "" { errs["name"] = "wajib diisi" }
	if req.NIM == "" { errs["nim"] = "wajib diisi" }
	
	// ngecek duplikasi nim yang kalo ada yang sama bakal ditolak
	for _, s := range students {
		if s.NIM == req.NIM {
			return fail(c, fiber.StatusConflict, "NIM sudah terdaftar")
		}
	}

	if len(errs) > 0 { return failValidation(c, errs) }

	baru := Student{
		ID:       nextID,
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: true,
	}
	students = append(students, baru)
	nextID++

	return created(c, "mahasiswa berhasil ditambah", baru, "/api/v1/students/"+strconv.Itoa(baru.ID))
}

// PUT /api/v1/students/:id
func replaceStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid { return fail(c, fiber.StatusBadRequest, "id harus angka") }

	i := findStudentIndex(id)
	if i == -1 { return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan") }

	var req ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus JSON")
	}

	errs := map[string]string{}
	if strings.TrimSpace(req.Name) == "" { errs["name"] = "wajib diisi pada PUT" }
	if strings.TrimSpace(req.NIM) == "" { errs["nim"] = "wajib diisi pada PUT" }
	if len(errs) > 0 { return failValidation(c, errs) }

	students[i].NIM = req.NIM
	students[i].Name = req.Name
	students[i].Grade = req.Grade
	students[i].IsActive = req.IsActive

	return ok(c, "data mahasiswa berhasil diganti seluruhnya", students[i])
}

// PATCH /api/v1/students/:id
func patchStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid { return fail(c, fiber.StatusBadRequest, "id harus angka") }

	i := findStudentIndex(id)
	if i == -1 { return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan") }

	var req PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus JSON")
	}

	if req.NIM != nil { students[i].NIM = *req.NIM }
	if req.Name != nil { students[i].Name = *req.Name }
	if req.Grade != nil { students[i].Grade = *req.Grade }
	if req.IsActive != nil { students[i].IsActive = *req.IsActive }

	return ok(c, "data mahasiswa berhasil diperbarui sebagian", students[i])
}

// DELETE /api/v1/students/:id
func deleteStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid { return fail(c, fiber.StatusBadRequest, "id harus angka") }

	i := findStudentIndex(id)
	if i == -1 { return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan") }

	students = append(students[:i], students[i+1:]...)
	return noContent(c)
}