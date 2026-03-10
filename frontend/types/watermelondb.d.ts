declare module '@nozbe/watermelondb' {
  export class Model {
    static table: string;
  }
  export class Database {
    constructor(options: { adapter: any; modelClasses: any[] });
    get(tableName: string): any;
    write(action: () => Promise<void>): Promise<void>;
  }
  export function appSchema(options: any): any;
  export function tableSchema(options: any): any;
}

declare module '@nozbe/watermelondb/adapters/sqlite' {
  interface SQLiteAdapterOptions {
    schema: any;
    jsi?: boolean;
    onSetUpError?: (error: any) => void;
  }
  export default class SQLiteAdapter {
    constructor(options: SQLiteAdapterOptions);
  }
}

declare module '@nozbe/watermelondb/decorators' {
  export function field(columnName: string): PropertyDecorator;
  export function date(columnName: string): PropertyDecorator;
  export function relation(tableName: string, relationIdColumn: string): PropertyDecorator;
  export function children(tableName: string): PropertyDecorator;
  export function readonly(decorator: PropertyDecorator): PropertyDecorator;
  export function nochange(decorator: PropertyDecorator): PropertyDecorator;
  export function json(columnName: string, sanitze: (raw: any) => any): PropertyDecorator;
  export function text(columnName: string): PropertyDecorator;
}

