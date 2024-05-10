import {
	array,
	boolean,
	merge,
	number,
	object,
	optional,
	picklist,
	string,
	type Output
} from 'valibot';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Scan Status
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const ScanStatusSchema = picklist(['waiting', 'processing', '']);
export type ScanStatus = Output<typeof ScanStatusSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const BaseSchema = object({
	id: string(),
	createdAt: string(),
	updatedAt: string()
});

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Attachment
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const AttachmentSchema = merge([
	BaseSchema,
	object({
		courseId: string(),
		assetId: string(),
		title: string(),
		path: string()
	})
]);

export type Attachment = Output<typeof AttachmentSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Asset
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const AssetTypeSchema = picklist(['video', 'html', 'pdf']);
export type AssetType = Output<typeof AssetTypeSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const AssetSchema = merge([
	BaseSchema,
	object({
		courseId: string(),
		title: string(),
		prefix: number(),
		chapter: string(),
		path: string(),
		assetType: AssetTypeSchema,

		// Progress
		videoPos: number(),
		completed: boolean(),
		completedAt: string(),

		// Attachments
		attachments: optional(array(AttachmentSchema))
	})
]);

export type Asset = Output<typeof AssetSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type AssetsGetParams = {
	orderBy?: string;
	page?: number;
	perPage?: number;
	expand?: boolean;
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Course
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const CourseSchema = merge([
	BaseSchema,
	object({
		title: string(),
		path: string(),
		hasCard: boolean(),
		available: boolean(),

		// Scan status
		scanStatus: ScanStatusSchema,

		// Progress
		started: boolean(),
		startedAt: string(),
		percent: number(),
		completedAt: string(),
		progressUpdatedAt: string()
	})
]);

export type Course = Output<typeof CourseSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export enum CourseProgress {
	NotStarted = 'Not Started',
	Started = 'Started',
	NotCompleted = 'Not Completed',
	Completed = 'Completed'
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type CoursesGetParams = {
	orderBy?: string;
	progress?: CourseProgress;
	tags?: string;
	page?: number;
	perPage?: number;
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type CoursePostParams = {
	title: string;
	path: string;
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type CourseChapters = Record<string, Asset[]>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Course Tags
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const CourseTagSchema = object({
	id: string(),
	tag: string(),
	forDeletion: optional(boolean())
});

export type CourseTag = Output<typeof CourseTagSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Tags
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const TagSchema = merge([
	BaseSchema,
	object({
		tag: string(),
		courseCount: number(),
		courses: optional(
			array(
				object({
					id: string(),
					title: string()
				})
			)
		)
	})
]);

export type Tag = Output<typeof TagSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type TagGetParams = {
	byName?: boolean;
	insensitive?: boolean;
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type TagsGetParams = {
	orderBy?: string;
	page?: number;
	perPage?: number;
	filter?: string;
	expand?: boolean;
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Scan
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const ScanSchema = merge([
	BaseSchema,
	object({
		courseId: string(),
		status: ScanStatusSchema
	})
]);

export type Scan = Output<typeof ScanSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type ScanPostParams = {
	courseId: string;
};
