package admin

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type developmentDepartment struct {
	code        string
	name        string
	description string
}

type developmentLocation struct {
	code        string
	name        string
	description string
}

const developmentDepartmentDescription = "Development department seed data"
const development96VillasLocationDescription = "Development location seed data: 96 Villas"
const developmentBWPLocationDescription = "Development location seed data: BWP"
const developmentBWPRoomsLocationDescription = "Development location seed data: BWP Rooms"

var developmentDepartments = []developmentDepartment{
	{code: "CON", name: "Concierge", description: developmentDepartmentDescription},
	{code: "DA", name: "Damaged Asset", description: developmentDepartmentDescription},
	{code: "FB", name: "F&B", description: developmentDepartmentDescription},
	{code: "FIN", name: "Finance Request", description: developmentDepartmentDescription},
	{code: "FO", name: "Front Office", description: developmentDepartmentDescription},
	{code: "HK", name: "Housekeeping", description: developmentDepartmentDescription},
	{code: "HKPPM", name: "Housekeeping PPM", description: developmentDepartmentDescription},
	{code: "IT", name: "IT", description: developmentDepartmentDescription},
	{code: "KIT", name: "Kitchen", description: developmentDepartmentDescription},
	{code: "LDRY", name: "Laundry", description: developmentDepartmentDescription},
	{code: "LF", name: "Lost & Found", description: developmentDepartmentDescription},
	{code: "MAINT", name: "Maintenance", description: developmentDepartmentDescription},
	{code: "REC", name: "REC", description: developmentDepartmentDescription},
	{code: "SEC", name: "Security", description: developmentDepartmentDescription},
}

var developmentLocations = []developmentLocation{
	{code: "LOBBY", name: "Lobby", description: "Development location seed data"},
	{code: "BALLROOM", name: "Ballroom", description: "Development location seed data"},
	{code: "BACK-OFFICE", name: "Back Office", description: "Development location seed data"},
	{code: "ROOM-8020", name: "Room 8020", description: "Development location seed data"},
	{code: "ROOM-7309", name: "Room 7309", description: "Development location seed data"},
	{code: "VILLA", name: "Villa", description: "Development location seed data"},
}

var development96VillasLocations = []developmentLocation{
	{code: "96V", name: "96 Villas", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-001", name: "96-BWV - Asian kitchen", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-002", name: "96-BWV - Auxiliary Swimming Pool", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-003", name: "96-BWV - Bathroom", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-004", name: "96-BWV - Buffet counter", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-005", name: "96-BWV - Cold kitchen", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-006", name: "96-BWV - Eng Fire Pump Room", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-007", name: "96-BWV - Eng Mainpool MEP Room", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-008", name: "96-BWV - Eng MEP Room", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-009", name: "96-BWV - Eng Subpool MEP Room", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-010", name: "96-BWV - Eng Water Treatment Room", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-011", name: "96-BWV - Eng Well Water treatment Room", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-012", name: "96-BWV - Eng workshop", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-013", name: "96-BWV - European kitchen", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-014", name: "96-BWV - Extra Pool", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-015", name: "96-BWV - Female Locker", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-016", name: "96-BWV - FO Reception", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-017", name: "96-BWV - Generator Room", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-018", name: "96-BWV - Gym", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-019", name: "96-BWV - HK Store", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-020", name: "96-BWV - Kid club", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-021", name: "96-BWV - Kid's Club", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-022", name: "96-BWV - Kid's Playground", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-023", name: "96-BWV - Lobby", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-024", name: "96-BWV - Lobby Lounge", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-025", name: "96-BWV - Main kitchen", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-026", name: "96-BWV - Main Pool", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-027", name: "96-BWV - Male Locker", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-028", name: "96-BWV - Outside", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-029", name: "96-BWV - PA Store", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-030", name: "96-BWV - Pastry kitchen", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-031", name: "96-BWV - Spa", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-032", name: "96-BWV - Toilet Gym", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-033", name: "96-BWV - Toilet Hồ bơi phụ", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-034", name: "96-BWV - Toilet Lobby", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-035", name: "96-BWV - Toilet Spa", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-036", name: "96-BWV - Toilet Tropicana", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-037", name: "96-BWV - Tropicana Bar", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-038", name: "96-BWV - Tropicana Kitchen", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-039", name: "96-BWV - Tropicana Restaurant", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-040", name: "96-BWV-Kitchen Office", description: development96VillasLocationDescription},
	{code: "96BWV-AREA-041", name: "96-BWV-Steward", description: development96VillasLocationDescription},
}

var developmentBWPLocations = []developmentLocation{
	{code: "BWP-AREA-001", name: "BOD Office", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-002", name: "BWP - Asian kitchen", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-003", name: "BWP - Back Office/FO", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-004", name: "BWP - Bakery kitchen", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-005", name: "BWP - Ballroom", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-006", name: "BWP - Basemant cold storage", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-007", name: "BWP - Beach Bar", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-008", name: "BWP - Bell Desk", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-009", name: "BWP - BOH Essence", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-010", name: "BWP - Boiler room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-011", name: "BWP - BTS room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-012", name: "BWP - Buffet counter", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-013", name: "BWP - Canteen", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-014", name: "BWP - Capentry work", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-015", name: "BWP - cold kitchen", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-016", name: "BWP - Cooling tower", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-017", name: "BWP - Cview Bar", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-018", name: "BWP - Cview Kitchen", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-019", name: "BWP - Eng store 2B", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-020", name: "BWP - Eng store 2C", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-021", name: "BWP - Eng store 3B", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-022", name: "BWP - Eng store 3C", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-023", name: "BWP - Eng store 4B", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-024", name: "BWP - Eng store 4C", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-025", name: "BWP - Eng store 5B", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-026", name: "BWP - Eng store 5C", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-027", name: "BWP - Eng store 6B", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-028", name: "BWP - Eng store 6C", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-029", name: "BWP - Eng store 7B", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-030", name: "BWP - Eng store 7C", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-031", name: "BWP - Eng store 8B", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-032", name: "BWP - Eng store 8C", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-033", name: "BWP - Eng store 9B", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-034", name: "BWP - Eng store 9C", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-035", name: "BWP - EPS room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-036", name: "BWP - Essence Kitchen", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-037", name: "BWP - Essence restaurant", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-038", name: "BWP - Essences Bar", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-039", name: "BWP - European kitchen", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-040", name: "BWP - Fan room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-041", name: "BWP - FB Office", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-042", name: "BWP - Female Locker", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-043", name: "BWP - FO Pantry", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-044", name: "BWP - FO Reception", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-045", name: "BWP - Generator room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-046", name: "BWP - GRO Counter", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-047", name: "BWP - Guest elevator wing 1", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-048", name: "BWP - Guest elevator wing 3", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-049", name: "BWP - Gym", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-050", name: "BWP - HK OFFICE", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-051", name: "BWP - HK Pantry Wing 2-2nd Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-052", name: "BWP - HK Pantry Wing 2-2nd Floorr", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-053", name: "BWP - HK Pantry Wing 2-3nd Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-054", name: "BWP - HK Pantry Wing 2-4th Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-055", name: "BWP - HK Pantry Wing 2-5th Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-056", name: "BWP - HK Pantry Wing 2-6th Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-057", name: "BWP - HK Pantry Wing 2-7th Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-058", name: "BWP - HK Pantry Wing 2-8th Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-059", name: "BWP - HK Pantry Wing 2-9th Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-060", name: "BWP - HK Pantry Wing 3-3th Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-061", name: "BWP - HK Pantry Wing 3-4th Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-062", name: "BWP - HK Pantry Wing 3-5th Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-063", name: "BWP - HK Pantry Wing 3-6th Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-064", name: "BWP - HK Pantry Wing 3-7th Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-065", name: "BWP - HK Pantry Wing 3-8th Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-066", name: "BWP - HK Pantry Wing 3-9th Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-067", name: "BWP - HK Store 2A1", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-068", name: "BWP - HK Store 2A3", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-069", name: "BWP - HK Store 3A1", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-070", name: "BWP - HK Store 3A3", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-071", name: "BWP - HK Store 4A1", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-072", name: "BWP - HK Store 4A3", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-073", name: "BWP - HK Store 5A1", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-074", name: "BWP - HK Store 5A3", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-075", name: "BWP - HK Store 6A1", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-076", name: "BWP - HK Store 6A3", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-077", name: "BWP - HK Store 7A1", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-078", name: "BWP - HK Store 7A3", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-079", name: "BWP - HK Store 8A1", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-080", name: "BWP - HK Store 8A3", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-081", name: "BWP - HK Store 9A1", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-082", name: "BWP - HK Store 9A3", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-083", name: "BWP - Ice machine room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-084", name: "BWP - In front of Ballroom", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-085", name: "BWP - Kid's Club", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-086", name: "BWP - Kid's Playground", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-087", name: "BWP - Kitchen office", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-088", name: "BWP - Lagoon", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-089", name: "BWP - Lagoon pump room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-090", name: "BWP - Lagoon Swimming Pool", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-091", name: "BWP - Laundry room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-092", name: "BWP - Lobby", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-093", name: "BWP - Main kitchen", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-094", name: "BWP - Main Swimming Pool", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-095", name: "BWP - Mainpool MEP Room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-096", name: "BWP - Male Locker", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-097", name: "BWP - Medium voltage room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-098", name: "BWP - Meeting room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-099", name: "BWP - Oasis bar", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-100", name: "BWP - Oasis Bathroom", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-101", name: "BWP - Oasis pool", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-102", name: "BWP - Oasis pool bar", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-103", name: "BWP - Oasis Swimming Pool", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-104", name: "BWP - Operator Room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-105", name: "BWP - Outside Lobby", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-106", name: "BWP - PA Store-1st Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-107", name: "BWP - PA Store-M Floor", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-108", name: "BWP - Pastry kitchen.", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-109", name: "BWP - PS/DS room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-110", name: "BWP - Pump filter room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-111", name: "BWP - Recreation Store", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-112", name: "BWP - RES Office", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-113", name: "BWP - Romantic Dinner", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-114", name: "BWP - Rooftop Fan", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-115", name: "BWP - Spa", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-116", name: "BWP - Staff restroom", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-117", name: "BWP - Staffhouse 1", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-118", name: "BWP - Staffhouse 2", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-119", name: "BWP - Toilet Ballroom", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-120", name: "BWP - Toilet Basement", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-121", name: "BWP - Toilet C view", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-122", name: "BWP - Toilet Essence", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-123", name: "BWP - Toilet Lobby", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-124", name: "BWP - Toilet Oasis", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-125", name: "BWP - Toilet Staff leve 1", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-126", name: "BWP - Toilet Staff leve M", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-127", name: "BWP - Transformer room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-128", name: "BWP - Uniform room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-129", name: "BWP - Wastewater treatmant room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-130", name: "BWP - Workshop", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-131", name: "BWP (16 Villas & B- 1st floor)", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-132", name: "BWP Codotel (M- Rooftop floor)", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-133", name: "BWP- MSB room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-134", name: "BWP- Oasis Kitchen", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-135", name: "BWP-Butchery Area", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-136", name: "BWP-Kitchen Office", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-137", name: "BWP-Receiving Area", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-138", name: "BWP-Recreation Office", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-139", name: "BWP-Steward Area", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-140", name: "CCTV Room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-141", name: "ENG-OFFICE", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-142", name: "Executive Office", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-143", name: "FIN DOCUMENT STORE", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-144", name: "FIN OFFICE", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-145", name: "GENERAL STORE", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-146", name: "HR Office", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-147", name: "IT Office", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-148", name: "MasterKey Audit", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-149", name: "Nurse Office", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-150", name: "RECEIVING OFFICE", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-151", name: "Sales Room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-152", name: "Security Office", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-153", name: "Server Room", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-154", name: "Server Room- 96 Villas", description: developmentBWPLocationDescription},
	{code: "BWP-AREA-155", name: "Staffhouse", description: developmentBWPLocationDescription},
}

var developmentBWPRoomNumbers = []int{
	// 2xxx: 72
	2001, 2002, 2003, 2004, 2005, 2006, 2007, 2008, 2009, 2010, 2011, 2012, 2014, 2015, 2016, 2017, 2018, 2020, 2022, 2024, 2026, 2028,
	2100, 2101, 2102, 2103, 2104, 2105, 2106, 2107, 2108, 2109, 2110, 2111, 2112, 2114, 2115, 2116, 2117, 2118, 2120, 2122, 2124, 2126,
	2200, 2201, 2202, 2203, 2204, 2205, 2206, 2207, 2208, 2209, 2211, 2215, 2217,
	2300, 2301, 2302, 2303, 2304, 2305, 2306, 2307, 2308, 2309, 2310, 2311, 2315, 2317, 2319,

	// 3xxx: 75
	3001, 3002, 3003, 3004, 3005, 3006, 3007, 3008, 3009, 3010, 3011, 3012, 3014, 3015, 3016, 3017, 3018, 3020, 3022, 3024, 3026, 3028,
	3100, 3101, 3102, 3103, 3104, 3105, 3106, 3107, 3108, 3109, 3110, 3111, 3112, 3114, 3115, 3116, 3117, 3118, 3120, 3122, 3124, 3126,
	3200, 3201, 3202, 3203, 3204, 3205, 3206, 3207, 3208, 3209, 3211, 3215, 3217, 3219, 3221,
	3300, 3301, 3302, 3303, 3304, 3305, 3306, 3307, 3308, 3309, 3310, 3311, 3315, 3317, 3319, 3321,

	// 4xxx: 75
	4001, 4002, 4003, 4004, 4005, 4006, 4007, 4008, 4009, 4010, 4011, 4012, 4014, 4015, 4016, 4017, 4018, 4020, 4022, 4024, 4026, 4028,
	4100, 4101, 4102, 4103, 4104, 4105, 4106, 4107, 4108, 4109, 4110, 4111, 4112, 4114, 4115, 4116, 4117, 4118, 4120, 4122, 4124, 4126,
	4200, 4201, 4202, 4203, 4204, 4205, 4206, 4207, 4208, 4209, 4211, 4215, 4217, 4219, 4221,
	4300, 4301, 4302, 4303, 4304, 4305, 4306, 4307, 4308, 4309, 4310, 4311, 4315, 4317, 4319, 4321,

	// 5xxx: 65
	5001, 5002, 5003, 5004, 5005, 5006, 5007, 5008, 5009, 5010, 5011, 5012, 5014, 5015, 5016, 5017, 5018, 5020, 5022, 5024, 5026, 5028,
	5100, 5101, 5102, 5103, 5104, 5105, 5106, 5107, 5108, 5109, 5110, 5111, 5112, 5114, 5115, 5116, 5117, 5118, 5120, 5122, 5124,
	5209, 5211, 5215, 5217, 5219, 5221,
	5300, 5301, 5302, 5303, 5304, 5305, 5306, 5307, 5308, 5309, 5310, 5311, 5315, 5317, 5319, 5321,

	// 6xxx: 60
	6001, 6002, 6003, 6004, 6005, 6006, 6007, 6008, 6009, 6010, 6011, 6012, 6014, 6015, 6016, 6017, 6019,
	6100, 6101, 6102, 6103, 6104, 6105, 6106, 6107, 6108, 6109, 6110, 6111, 6112, 6114, 6115, 6117, 6119,
	6200, 6201, 6202, 6203, 6205, 6207, 6209, 6211, 6215, 6217, 6219, 6221,
	6300, 6301, 6302, 6303, 6304, 6305, 6307, 6309, 6311, 6315, 6317, 6319, 6321, 6666,

	// 7xxx: 75
	7001, 7002, 7003, 7004, 7005, 7006, 7007, 7008, 7009, 7010, 7011, 7012, 7014, 7015, 7016, 7017, 7018, 7019, 7020, 7022, 7024, 7026,
	7100, 7101, 7102, 7103, 7104, 7105, 7106, 7107, 7108, 7109, 7110, 7111, 7112, 7114, 7115, 7116, 7117, 7118, 7119, 7120, 7122, 7124,
	7200, 7201, 7202, 7203, 7204, 7205, 7206, 7207, 7208, 7209, 7211, 7215, 7217, 7219, 7221,
	7300, 7301, 7302, 7303, 7304, 7305, 7306, 7307, 7308, 7309, 7310, 7311, 7315, 7317, 7319, 7321,

	// 8xxx: 76
	8001, 8002, 8003, 8004, 8005, 8006, 8007, 8008, 8009, 8010, 8011, 8012, 8014, 8015, 8016, 8017, 8018, 8019, 8020, 8022, 8024, 8026,
	8100, 8101, 8102, 8103, 8104, 8105, 8106, 8107, 8108, 8109, 8110, 8111, 8112, 8114, 8115, 8116, 8117, 8118, 8119, 8120, 8122, 8124,
	8200, 8201, 8202, 8203, 8204, 8205, 8206, 8207, 8208, 8209, 8211, 8215, 8217, 8219, 8221,
	8300, 8301, 8302, 8303, 8304, 8305, 8306, 8307, 8308, 8309, 8310, 8311, 8315, 8317, 8319, 8321, 8888,

	// 9xxx: 66
	9001, 9002, 9003, 9004, 9005, 9006, 9007, 9008, 9009, 9010, 9011, 9012, 9014, 9015, 9017,
	9100, 9101, 9102, 9103, 9104, 9105, 9106, 9107, 9108, 9109, 9110, 9111,
	9200, 9201, 9202, 9203, 9205, 9207, 9209, 9211, 9215, 9217, 9219, 9221,
	9300, 9301, 9302, 9303, 9304, 9305, 9307, 9309, 9311, 9315, 9317, 9319, 9321,
	9501, 9502, 9503, 9504, 9505, 9506, 9507, 9508, 9509, 9510, 9511,
	9601, 9602, 9999,
}

// SeedDevelopmentFixtures inserts only local development reference data.
// Callers must enforce the development-only configuration guard.
func SeedDevelopmentFixtures(ctx context.Context, pool *pgxpool.Pool) error {
	transaction, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin development fixtures seed: %w", err)
	}

	for _, department := range developmentDepartments {
		if _, err := transaction.Exec(ctx, `
INSERT INTO departments (code, name, description)
VALUES ($1, $2, $3)
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    description = EXCLUDED.description,
    is_active = TRUE,
    updated_at = NOW()
WHERE departments.description = $3`, department.code, department.name, department.description); err != nil {
			_ = transaction.Rollback(ctx)
			return fmt.Errorf("insert development department %s: %w", department.code, err)
		}
	}

	if _, err := transaction.Exec(ctx, `
UPDATE departments
SET is_active = FALSE,
    updated_at = NOW()
WHERE code = ANY($1::text[])
  AND description = $2`, []string{"ENG", "HR"}, developmentDepartmentDescription); err != nil {
		_ = transaction.Rollback(ctx)
		return fmt.Errorf("deactivate legacy development departments: %w", err)
	}

	for _, location := range developmentLocations {
		if _, err := transaction.Exec(ctx, `
INSERT INTO locations (code, name, description)
VALUES ($1, $2, $3)
ON CONFLICT (code) DO NOTHING`, location.code, location.name, location.description); err != nil {
			_ = transaction.Rollback(ctx)
			return fmt.Errorf("insert development location %s: %w", location.code, err)
		}
	}

	if _, err := transaction.Exec(ctx, `
UPDATE locations
SET is_active = FALSE,
    updated_at = NOW()
WHERE code = ANY($1::text[])
  AND description = $2`, []string{"ROOM-8020", "ROOM-7309"}, "Development location seed data"); err != nil {
		_ = transaction.Rollback(ctx)
		return fmt.Errorf("deactivate obsolete development locations: %w", err)
	}

	for _, location := range development96VillasLocations {
		if err := seed96VillasLocation(ctx, transaction, location); err != nil {
			_ = transaction.Rollback(ctx)
			return fmt.Errorf("seed 96 Villas location %s: %w", location.code, err)
		}
	}
	for room := 1001; room <= 1099; room++ {
		location := developmentLocation{
			code:        fmt.Sprintf("96BWV-ROOM-%d", room),
			name:        fmt.Sprintf("%d", room),
			description: development96VillasLocationDescription,
		}
		if err := seed96VillasLocation(ctx, transaction, location); err != nil {
			_ = transaction.Rollback(ctx)
			return fmt.Errorf("seed 96 Villas room %d: %w", room, err)
		}
	}
	for _, location := range developmentBWPLocations {
		if err := seedBWPLocation(ctx, transaction, location); err != nil {
			_ = transaction.Rollback(ctx)
			return fmt.Errorf("seed BWP location %s: %w", location.code, err)
		}
	}
	for _, roomNumber := range developmentBWPRoomNumbers {
		name := fmt.Sprintf("%d", roomNumber)
		location := developmentLocation{
			code:        "BWP-ROOM-" + name,
			name:        name,
			description: developmentBWPRoomsLocationDescription,
		}
		if err := seedBWPRoomLocation(ctx, transaction, location); err != nil {
			_ = transaction.Rollback(ctx)
			return fmt.Errorf("seed BWP room %d: %w", roomNumber, err)
		}
	}

	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("commit development fixtures seed: %w", err)
	}
	return nil
}

func seed96VillasLocation(ctx context.Context, transaction pgx.Tx, location developmentLocation) error {
	return seedMarkedLocation(ctx, transaction, location, "96 Villas")
}

func seedBWPLocation(ctx context.Context, transaction pgx.Tx, location developmentLocation) error {
	return seedMarkedLocation(ctx, transaction, location, "BWP")
}

func seedBWPRoomLocation(ctx context.Context, transaction pgx.Tx, location developmentLocation) error {
	return seedMarkedLocation(ctx, transaction, location, "BWP Rooms")
}

func seedMarkedLocation(ctx context.Context, transaction pgx.Tx, location developmentLocation, fixtureName string) error {
	result, err := transaction.Exec(ctx, `
INSERT INTO locations (code, name, description, is_active)
VALUES ($1, $2, $3, TRUE)
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    description = EXCLUDED.description,
    is_active = TRUE,
    updated_at = NOW()
WHERE locations.description = $3`, location.code, location.name, location.description)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("%s location code %q collides with non-fixture row", fixtureName, location.code)
	}
	return nil
}
