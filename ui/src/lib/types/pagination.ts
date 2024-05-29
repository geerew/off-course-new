import { array, number, object, union, type Output } from 'valibot';
import { AssetSchema, CourseSchema, LogSchema, TagSchema } from './models';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type PaginationParams = {
	page: number;
	perPage: number;
	perPages: number[];
	totalItems: number;
	totalPages: number;
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const PaginationSchema = object({
	page: number(),
	perPage: number(),
	totalItems: number(),
	totalPages: number(),
	items: union([array(CourseSchema), array(AssetSchema), array(TagSchema), array(LogSchema)])
});

export type Pagination = Output<typeof PaginationSchema>;
