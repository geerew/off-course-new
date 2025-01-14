import {
	any,
	array,
	boolean,
	number,
	object,
	optional,
	picklist,
	record,
	string,
	type InferOutput
} from 'valibot';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Scan Status
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const ScanStatusSchema = picklist(['waiting', 'processing', '']);
export type ScanStatus = InferOutput<typeof ScanStatusSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const BaseSchema = object({
	id: string(),
	createdAt: string(),
	updatedAt: string()
});

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Attachment
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const AttachmentSchema = object({
	...BaseSchema.entries,
	...object({
		courseId: string(),
		assetId: string(),
		title: string(),
		path: string()
	}).entries
});

export type Attachment = InferOutput<typeof AttachmentSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Asset
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const AssetTypeSchema = picklist(['video', 'html', 'pdf']);
export type AssetType = InferOutput<typeof AssetTypeSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const AssetSchema = object({
	...BaseSchema.entries,
	...object({
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
	}).entries
});

export type Asset = InferOutput<typeof AssetSchema>;

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

export const CourseSchema = object({
	...BaseSchema.entries,
	...object({
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
	}).entries
});

export type Course = InferOutput<typeof CourseSchema>;

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
	titles?: string;
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

export type CourseTag = InferOutput<typeof CourseTagSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Tags
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const TagSchema = object({
	...BaseSchema.entries,
	...object({
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
	}).entries
});

export type Tag = InferOutput<typeof TagSchema>;

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

export const ScanSchema = object({
	...BaseSchema.entries,
	...object({
		courseId: string(),
		status: ScanStatusSchema
	}).entries
});

export type Scan = InferOutput<typeof ScanSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type ScanPostParams = {
	courseId: string;
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Log
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const LogSchema = object({
	id: string(),
	level: number(),
	message: string(),
	data: record(string(), any()),
	createdAt: string()
});

export type Log = InferOutput<typeof LogSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type LogsGetParams = {
	messages?: string;
	levels?: string;
	types?: string;
	page?: number;
	perPage?: number;
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export enum LogLevel {
	DEBUG = 'debug',
	INFO = 'info',
	WARN = 'warn',
	ERROR = 'error'
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const LogLevelMapping: { [key in LogLevel]: number } = {
	[LogLevel.DEBUG]: -4,
	[LogLevel.INFO]: 0,
	[LogLevel.WARN]: 4,
	[LogLevel.ERROR]: 8
};
