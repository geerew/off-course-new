import { array, boolean, enumType, object, optional, string, type Output } from 'valibot';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Scan Status
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const ScanStatusSchema = enumType(['waiting', 'processing', '']);
export type ScanStatus = Output<typeof ScanStatusSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Asset
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const AssetTypeSchema = enumType(['video', 'html', 'pdf']);
export type AssetType = Output<typeof AssetTypeSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const AssetSchema = object({
	id: string(),
	courseId: string(),
	title: string(),
	prefix: string(),
	chapter: string(),
	path: string(),
	assetType: AssetTypeSchema,
	started: boolean(),
	finished: boolean(),
	createdAt: string(),
	updatedAt: string()
});

export type Asset = Output<typeof AssetSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Course
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const CourseSchema = object({
	id: string(),
	title: string(),
	path: string(),
	hasCard: boolean(),
	started: boolean(),
	finished: boolean(),
	scanStatus: ScanStatusSchema,
	createdAt: string(),
	updatedAt: string(),

	// Relations
	assets: optional(array(AssetSchema))
});

export type Course = Output<typeof CourseSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type CoursesGetParams = {
	orderBy?: string;
	includeAssets?: boolean;
	page?: number;
	perPage?: number;
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type CourseGetParams = {
	includeAssets?: boolean;
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type CoursePostParams = {
	title: string;
	path: string;
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type CourseChapters = Record<string, Asset[]>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Scan
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const ScanSchema = object({
	id: string(),
	courseId: string(),
	status: ScanStatusSchema,
	createdAt: string(),
	updatedAt: string()
});

export type Scan = Output<typeof ScanSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type ScanPostParams = {
	courseId: string;
};
