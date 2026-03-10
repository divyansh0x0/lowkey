import { Database } from '@nozbe/watermelondb';
import SQLiteAdapter from '@nozbe/watermelondb/adapters/sqlite';
import { schema } from '../model/schema';
import { Message } from '../model/Message';

const adapter = new SQLiteAdapter({
  schema,
  jsi: true,
  onSetUpError: (error: any) => {
    console.error('WatermelonDB setup error:', error);
  },
});

export const database = new Database({
  adapter,
  modelClasses: [Message],
});
